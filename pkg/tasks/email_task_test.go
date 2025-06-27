package tasks

import (
	"context"
	"testing"
	"time"

	"github.com/mikestefanello/pagoda/config"
	"github.com/mikestefanello/pagoda/pkg/services"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailWorker_Work(t *testing.T) {
	// Setup: Create a minimal service container for the worker
	// In a real scenario, you might mock parts of this, especially c.Mail if it made external calls.
	cfg := config.Config{
		App: config.AppConfig{
			Environment: config.EnvTest, // Ensures mailer might skip actual sending if configured
		},
		Mail: config.MailConfig{
			FromAddress: "test@example.com",
		},
	}

	// We need a real container to pass to NewEmailWorker as it's currently structured.
	// For a more isolated unit test, EmailWorker could take an interface for dependencies.
	mockContainer := &services.Container{
		Config: &cfg,
	}

	// Initialize mail client as EmailWorker might use it (even if just for logging via container)
	// If NewMailClient could fail or needs more from cfg, this would need adjustment.
	var err error
	mockContainer.Mail, err = services.NewMailClient(&cfg)
	require.NoError(t, err)

	// Create the worker instance
	worker := NewEmailWorker(mockContainer)

	// Prepare job arguments
	args := EmailArgs{
		UserID:       123,
		EmailAddress: "recipient@example.com",
		Subject:      "Test Email",
		Body:         "Hello, this is a test email!",
	}

	// Prepare a River job
	// For unit testing Work, we don't need a full River client, just the job struct.
	// Some fields of river.Job are not relevant for direct Work() testing if not used by the worker.
	riverJob := &river.Job[EmailArgs]{
		Args:       args,
		ID:         "test-job-id",
		Kind:       args.Kind(),
		Queue:      river.QueueDefault,
		NumRetries: 0,
		RecordedAt: time.Now(),
		// Other fields like State, AttemptedAt, MaxAttempts could be set if the worker uses them.
	}

	// Execute the Work method
	err = worker.Work(context.Background(), riverJob)

	// Assertions
	assert.NoError(t, err, "EmailWorker.Work() should not return an error for successful processing")

	// Further assertions could be made here if the worker had more side effects,
	// e.g., checking if a mock mailer's Send method was called.
	// Since our current worker mainly logs, we could also capture and verify log output
	// by configuring a test-specific slog.Handler.
}
