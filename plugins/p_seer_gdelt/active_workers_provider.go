package p_seer_gdelt

import (
	"log"
	"log/slog"
	"time"

	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
	"gorm.io/gorm"
)

type gdeltWorkerInstance struct {
	name     string
	interval time.Duration
	lastRun  *time.Time
	nextRun  *time.Time
}

func (w *gdeltWorkerInstance) Name() string            { return w.name }
func (w *gdeltWorkerInstance) Interval() time.Duration { return w.interval }
func (w *gdeltWorkerInstance) LastRun() *time.Time     { return w.lastRun }
func (w *gdeltWorkerInstance) NextRun() *time.Time     { return w.nextRun }

type gdeltActiveWorkersProvider struct{}

func (gdeltActiveWorkersProvider) FetchActiveWorkers(db *gorm.DB) []p_seer_workerregistry.WorkerInstance {
	if db == nil {
		return nil
	}
	var workers []GDELTWorker
	if err := db.Order("name ASC").Find(&workers).Error; err != nil {
		return nil
	}
	out := make([]p_seer_workerregistry.WorkerInstance, 0, len(workers))
	for i := range workers {
		w := workers[i]
		lastLog, err := p_seer_workerregistry.LatestWorkerRunLog(db, p_seer_workerregistry.WorkerRunnerKindGDELT, w.ID)
		if err != nil {
			slog.Error("p_seer_gdelt: active workers latest run log", "error", err, "worker_id", w.ID)
			lastLog = nil
		}
		var lastRun, nextRun *time.Time
		if lastLog != nil && lastLog.FinishedAt != nil {
			t := lastLog.FinishedAt.UTC()
			lastRun = &t
			if w.Duration > 0 {
				n := t.Add(w.Duration)
				nextRun = &n
			}
		}
		out = append(out, &gdeltWorkerInstance{
			name:     w.Name,
			interval: w.Duration,
			lastRun:  lastRun,
			nextRun:  nextRun,
		})
	}
	return out
}

func init() {
	if err := p_seer_workerregistry.RegistryActiveWorkersProvider.Register("GDELT", gdeltActiveWorkersProvider{}); err != nil {
		log.Panic(err)
	}
}
