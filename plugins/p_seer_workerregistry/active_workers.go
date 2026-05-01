package p_seer_workerregistry

import (
	"time"

	"github.com/UniquityVentures/lago/registry"
	"gorm.io/gorm"
)

// WorkerInstance describes one scheduled worker row for dashboard display.
type WorkerInstance interface {
	Name() string
	LastRun() *time.Time
	NextRun() *time.Time
	Interval() time.Duration
}

// ActiveWorkersProvider returns worker rows for one tab (e.g. Reddit, Website).
type ActiveWorkersProvider interface {
	FetchActiveWorkers(db *gorm.DB) []WorkerInstance
}

// RegistryActiveWorkersProvider maps tab label -> provider (e.g. "Reddit", "Website").
var RegistryActiveWorkersProvider = registry.NewRegistry[ActiveWorkersProvider]()
