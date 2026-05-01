package p_seer_websites

import (
	"log"
	"log/slog"
	"time"

	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
	"gorm.io/gorm"
)

type websiteWorkerInstance struct {
	name     string
	interval time.Duration
	lastRun  *time.Time
	nextRun  *time.Time
}

func (w *websiteWorkerInstance) Name() string             { return w.name }
func (w *websiteWorkerInstance) Interval() time.Duration { return w.interval }
func (w *websiteWorkerInstance) LastRun() *time.Time      { return w.lastRun }
func (w *websiteWorkerInstance) NextRun() *time.Time      { return w.nextRun }

type websiteActiveWorkersProvider struct{}

func (websiteActiveWorkersProvider) FetchActiveWorkers(db *gorm.DB) []p_seer_workerregistry.WorkerInstance {
	if db == nil {
		return nil
	}
	var runners []WebsiteRunner
	if err := db.Order("name ASC").Find(&runners).Error; err != nil {
		return nil
	}
	out := make([]p_seer_workerregistry.WorkerInstance, 0, len(runners))
	for i := range runners {
		r := runners[i]
		lastLog, err := p_seer_workerregistry.LatestWorkerRunLog(db, p_seer_workerregistry.WorkerRunnerKindWebsite, r.ID)
		if err != nil {
			slog.Error("p_seer_websites: active workers latest run log", "error", err, "runner_id", r.ID)
			lastLog = nil
		}
		var lastRun, nextRun *time.Time
		if lastLog != nil && lastLog.FinishedAt != nil {
			t := lastLog.FinishedAt.UTC()
			lastRun = &t
			if r.Duration > 0 {
				n := t.Add(r.Duration)
				nextRun = &n
			}
		}
		out = append(out, &websiteWorkerInstance{
			name:     r.Name,
			interval: r.Duration,
			lastRun:  lastRun,
			nextRun:  nextRun,
		})
	}
	return out
}

func init() {
	if err := p_seer_workerregistry.RegistryActiveWorkersProvider.Register("Website", websiteActiveWorkersProvider{}); err != nil {
		log.Panic(err)
	}
}
