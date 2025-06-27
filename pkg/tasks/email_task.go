package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/mikestefanello/pagoda/pkg/log"
	"github.com/mikestefanello/pagoda/pkg/services"
)

// EmailArgs defines the arguments for our email sending job.
type EmailArgs struct {
	UserID       int    `json:"user_id"`
	EmailAddress string `json:"email_address"`
	Subject      string `json:"subject"`
	Body         string `json:"body"`
}

// Kind returns a string that uniquely identifies this type of job.
func (EmailArgs) Kind() string {
	return "send_email"
}

// EmailWorker is the worker for EmailArgs.
// It includes an injected MailClient (or the full service container) to perform its work.
type EmailWorker struct {
	river.WorkerDefaults[EmailArgs] // Embeds default methods for Worker
	Container *services.Container     // Service container for dependency injection
}

// Work performs the actual job of sending an email.
func (w *EmailWorker) Work(ctx context.Context, job *river.Job[EmailArgs]) error {
	log.Default().Info("Starting email task", "user_id", job.Args.UserID, "email", job.Args.EmailAddress)

	// In a real application, you would use the mail client:
	// err := w.Container.Mail.Send(job.Args.EmailAddress, job.Args.Subject, job.Args.Body)
	// if err != nil {
	// 	log.Default().Error("Failed to send email", "user_id", job.Args.UserID, "error", err)
	// 	return err // Returning an error will make River retry the job based on policy
	// }

	// Simulate work
	time.Sleep(2 * time.Second)

	log.Default().Info("Successfully processed email task",
		"user_id", job.Args.UserID,
		"email", job.Args.EmailAddress,
		"subject", job.Args.Subject,
	)
	return nil
}

// NewEmailWorker creates a new EmailWorker with its dependencies.
func NewEmailWorker(c *services.Container) *EmailWorker {
	return &EmailWorker{
		Container: c,
	}
}
