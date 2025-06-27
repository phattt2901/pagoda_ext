package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/mikestefanello/pagoda/pkg/handlers"
	"github.com/mikestefanello/pagoda/pkg/log"
	"github.com/mikestefanello/pagoda/pkg/services"
	"github.com/riverqueue/river" // Added River import
	// "github.com/mikestefanello/pagoda/pkg/tasks" // Removed backlite related task registration
)

func main() {
	// Start a new container.
	c := services.NewContainer()
	defer func() {
		// Gracefully shutdown all services.
		fatal("shutdown failed", c.Shutdown())
	}()

	// Build the router.
	if err := handlers.BuildRouter(c); err != nil {
		fatal("failed to build the router", err)
	}

	// Register all task queues.
	// tasks.Register(c) // This will be for River workers later

	// Migrate River schema
	if c.River != nil {
		log.Default().Info("Migrating River schema...")
		if err := c.River.Migrate(context.Background(), river.MigrationDirectionUp, nil); err != nil {
			fatal("failed to migrate River schema", err)
		}
		log.Default().Info("River schema migration complete.")

		// Start the River client to process jobs.
		// Workers need to be registered with the client before starting.
		// This will be handled in the pkg/tasks/register.go and called before starting.
		log.Default().Info("Starting River client...")
		if err := c.River.Start(context.Background()); err != nil {
			fatal("failed to start River client", err)
		}
		log.Default().Info("River client started.")
	}

	// Start the server.
	go func() {
		srv := http.Server{
			Addr:         fmt.Sprintf("%s:%d", c.Config.HTTP.Hostname, c.Config.HTTP.Port),
			Handler:      c.Web,
			ReadTimeout:  c.Config.HTTP.ReadTimeout,
			WriteTimeout: c.Config.HTTP.WriteTimeout,
			IdleTimeout:  c.Config.HTTP.IdleTimeout,
		}

		if c.Config.HTTP.TLS.Enabled {
			certs, err := tls.LoadX509KeyPair(c.Config.HTTP.TLS.Certificate, c.Config.HTTP.TLS.Key)
			fatal("cannot load TLS certificate", err)

			srv.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{certs},
			}
		}

		if err := c.Web.StartServer(&srv); errors.Is(err, http.ErrServerClosed) {
			fatal("shutting down the server", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the web server and task runner.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, os.Kill)
	<-quit
}

// fatal logs an error and terminates the application, if the error is not nil.
func fatal(msg string, err error) {
	if err != nil {
		log.Default().Error(msg, "error", err)
		os.Exit(1)
	}
}
