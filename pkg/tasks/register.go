package tasks

import (
	"github.com/mikestefanello/pagoda/pkg/services"
	"github.com/riverqueue/river"
)

// RegisterRiverWorkers registers all application River workers with the provided Workers instance.
func RegisterRiverWorkers(workers *river.Workers, c *services.Container) {
	// Create and register the EmailWorker
	emailWorker := NewEmailWorker(c)
	river.AddWorker(workers, emailWorker)

	// TODO: Register other workers here as they are created
	// For example:
	// anotherWorker := NewAnotherWorker(c)
	// river.AddWorker(workers, anotherWorker)
}
