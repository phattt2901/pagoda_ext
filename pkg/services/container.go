package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/labstack/echo/v4"
	_ "github.com/jackc/pgx/v5/stdlib" // Import pgx driver
	// "github.com/mikestefanello/backlite" // Removed backlite
	"github.com/mikestefanello/pagoda/config"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/pkg/log"
	"github.com/mikestefanello/pagoda/pkg/tasks" // Added tasks import for worker registration
	"github.com/riverqueue/river"               // Added River import
	"github.com/riverqueue/river/driver/riverpgxv5" // Added River pgx v5 driver
	"github.com/spf13/afero"

	// Required by ent.
	_ "github.com/mikestefanello/pagoda/ent/runtime"
)

// Container contains all services used by the application and provides an easy way to handle dependency
// injection including within tests.
type Container struct {
	// Validator stores a validator
	Validator *Validator

	// Web stores the web framework.
	Web *echo.Echo

	// Config stores the application configuration.
	Config *config.Config

	// Cache contains the cache client.
	Cache *CacheClient

	// Database stores the connection to the database.
	Database *sql.DB

	// Files stores the file system.
	Files afero.Fs

	// ORM stores a client to the ORM.
	ORM *ent.Client

	// Graph is the entity graph defined by your Ent schema.
	Graph *gen.Graph

	// Mail stores an email sending client.
	Mail *MailClient

	// Auth stores an authentication client.
	Auth *AuthClient

	// River stores the River client for task queueing.
	River *river.Client
}

// NewContainer creates and initializes a new Container.
func NewContainer() *Container {
	c := new(Container)
	c.initConfig()
	c.initValidator()
	c.initWeb()
	c.initCache()
	c.initDatabase()
	c.initFiles()
	c.initORM()
	c.initAuth()
	c.initMail()
	// c.initTasks() // Removed backlite
	c.initRiver()
	return c
}

// Shutdown gracefully shuts the Container down and disconnects all connections.
func (c *Container) Shutdown() error {
	// Shutdown the web server.
	webCtx, webCancel := context.WithTimeout(context.Background(), c.Config.HTTP.ShutdownTimeout)
	defer webCancel()
	if err := c.Web.Shutdown(webCtx); err != nil {
		return err
	}

	// Shutdown the task runner.
	// taskCtx, taskCancel := context.WithTimeout(context.Background(), c.Config.Tasks.ShutdownTimeout) // Removed backlite
	// defer taskCancel() // Removed backlite
	// c.Tasks.Stop(taskCtx) // Removed backlite

	// Shutdown River client
	if c.River != nil {
		// TODO: Determine appropriate timeout for River, for now using existing task shutdown timeout.
		// This might need adjustment based on typical job duration.
		riverCtx, riverCancel := context.WithTimeout(context.Background(), c.Config.Tasks.ShutdownTimeout)
		defer riverCancel()
		if err := c.River.Stop(riverCtx); err != nil {
			// Log the error but don't necessarily block other shutdowns
			log.Default().Error("failed to stop River client", "error", err)
		}
	}

	// Shutdown the ORM.
	if err := c.ORM.Close(); err != nil {
		return err
	}

	// Shutdown the database.
	if err := c.Database.Close(); err != nil {
		return err
	}

	// Shutdown the cache.
	c.Cache.Close()

	return nil
}

// initConfig initializes configuration.
func (c *Container) initConfig() {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	c.Config = &cfg

	// Configure logging.
	switch cfg.App.Environment {
	case config.EnvProduction:
		slog.SetLogLoggerLevel(slog.LevelInfo)
	default:
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

// initValidator initializes the validator.
func (c *Container) initValidator() {
	c.Validator = NewValidator()
}

// initWeb initializes the web framework.
func (c *Container) initWeb() {
	c.Web = echo.New()
	c.Web.HideBanner = true
	c.Web.Validator = c.Validator
}

// initCache initializes the cache.
func (c *Container) initCache() {
	store, err := newInMemoryCache(c.Config.Cache.Capacity)
	if err != nil {
		panic(err)
	}

	c.Cache = NewCacheClient(store)
}

// initDatabase initializes the database.
func (c *Container) initDatabase() {
	var err error
	var driverName string
	var connectionString string

	// TODO: The Driver string in config.yaml should be updated to "pgx" or "postgres"
	// For now, we will assume it's correctly set for PostgreSQL.
	// If you want to keep supporting SQLite, more sophisticated logic is needed here.
	driverName = "pgx" // Or c.Config.Database.Driver if it's set to "pgx" or "postgres" in config

	switch c.Config.App.Environment {
	case config.EnvTest:
		connectionString = c.Config.Database.PostgresTestDSN
		// Ensure TestConnection is set in your config for Postgres tests
		if connectionString == "" {
			panic("PostgresTestDSN is not set in config for test environment")
		}
	default:
		connectionString = c.Config.Database.PostgresDSN
		if connectionString == "" {
			panic("PostgresDSN is not set in config")
		}
	}

	c.Database, err = openDB(driverName, connectionString)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}
}

// initFiles initializes the file system.
func (c *Container) initFiles() {
	// Use in-memory storage for tests.
	if c.Config.App.Environment == config.EnvTest {
		c.Files = afero.NewMemMapFs()
		return
	}

	fs := afero.NewOsFs()
	if err := fs.MkdirAll(c.Config.Files.Directory, 0755); err != nil {
		panic(err)
	}
	c.Files = afero.NewBasePathFs(fs, c.Config.Files.Directory)
}

// initORM initializes the ORM.
func (c *Container) initORM() {
	// Ensure the driver name used here matches what's used in initDatabase,
	// or directly use the configured driver name if it's guaranteed to be for Postgres.
	driverName := "pgx" // Or c.Config.Database.Driver
	drv := entsql.OpenDB(driverName, c.Database)
	c.ORM = ent.NewClient(ent.Driver(drv))

	// Run the auto migration tool.
	if err := c.ORM.Schema.Create(context.Background()); err != nil {
		panic(err)
	}

	// Load the graph.
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	p := filepath.Join(filepath.Dir(d), "../ent/schema")
	g, err := entc.LoadGraph(p, &gen.Config{})
	if err != nil {
		panic(err)
	}
	c.Graph = g
}

// initAuth initializes the authentication client.
func (c *Container) initAuth() {
	c.Auth = NewAuthClient(c.Config, c.ORM)
}

// initMail initialize the mail client.
func (c *Container) initMail() {
	var err error
	c.Mail, err = NewMailClient(c.Config)
	if err != nil {
		panic(fmt.Sprintf("failed to create mail client: %v", err))
	}
}

/* Removed backlite initTasks
// initTasks initializes the task client.
func (c *Container) initTasks() {
	var err error
	// You could use a separate database for tasks, if you'd like, but using one
	// makes transaction support easier.
	c.Tasks, err = backlite.NewClient(backlite.ClientConfig{
		DB:              c.Database,
		Logger:          log.Default(),
		NumWorkers:      c.Config.Tasks.Goroutines,
		ReleaseAfter:    c.Config.Tasks.ReleaseAfter,
		CleanupInterval: c.Config.Tasks.CleanupInterval,
	})

	if err != nil {
		panic(fmt.Sprintf("failed to create task client: %v", err))
	}

	if err = c.Tasks.Install(); err != nil {
		panic(fmt.Sprintf("failed to install task schema: %v", err))
	}
}
*/

// initRiver initializes the River client.
func (c *Container) initRiver() {
	// Workers are registered here before the client is created.
	workers := river.NewWorkers()
	tasks.RegisterRiverWorkers(workers, c) // Pass the container 'c' for dependency injection into workers

	riverConfig := &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers: workers,
		Logger:  log.Default(), // Use the application's logger
	}

	// Use the existing *sql.DB from the container
	dbDriver := riverpgxv5.New(c.Database)

	var err error
	c.River, err = river.NewClient(dbDriver, riverConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create River client: %w", err))
	}

	// River client's Start() method will be called in cmd/web/main.go
	// River's database migrations (creating river_jobs table) are usually handled by calling
	// `riverClient.Migrate(ctx, river.MigrationDirectionUp, nil)`
	// This could be done here or as a separate startup step.
	// For now, let's assume it will be handled or the table will be created manually.
	// Alternatively, River might create it on first run if not present, depending on its internal logic.
	// The documentation suggests running migrations explicitly:
	// riverClient.Migrate(ctx, river.MigrationDirectionUp, &river.MigrateOpts{})
	// We will add this to the Start method in main.go or a dedicated setup function.
}

// openDB opens a database connection.
func openDB(driver, connection string) (*sql.DB, error) {
	// The SQLite specific logic for directory creation and $RAND can be removed
	// or conditionalized if you want to maintain dual DB support.
	// For this migration, we are focusing on PostgreSQL.
	if driver == "sqlite3" {
		// This block can be removed if SQLite is no longer supported.
		// Helper to automatically create the directories that the specified sqlite file
		// should reside in, if one.
		d := strings.Split(connection, "/")
		if len(d) > 1 {
			dirpath := strings.Join(d[:len(d)-1], "/")

			if err := os.MkdirAll(dirpath, 0755); err != nil {
				return nil, err
			}
		}

		// Check if a random value is required, which is often used for in-memory test databases.
		if strings.Contains(connection, "$RAND") {
			connection = strings.Replace(connection, "$RAND", fmt.Sprint(rand.Int()), 1)
		}
	}

	db, err := sql.Open(driver, connection)
	if err != nil {
		return nil, err
	}

	// Verify the connection to the database is established.
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Default().Info("Successfully connected to the database", "driver", driver)
	return db, nil
}
