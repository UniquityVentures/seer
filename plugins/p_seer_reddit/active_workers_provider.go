package p_seer_reddit

import (
	"log"
	"log/slog"
	"time"

	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
	"gorm.io/gorm"
)

type redditWorkerInstance struct {
	name     string
	interval time.Duration
	lastRun  *time.Time
	nextRun  *time.Time
}

func (w *redditWorkerInstance) Name() string               { return w.name }
func (w *redditWorkerInstance) Interval() time.Duration   { return w.interval }
func (w *redditWorkerInstance) LastRun() *time.Time       { return w.lastRun }
func (w *redditWorkerInstance) NextRun() *time.Time       { return w.nextRun }

type redditActiveWorkersProvider struct{}

func (redditActiveWorkersProvider) FetchActiveWorkers(db *gorm.DB) []p_seer_workerregistry.WorkerInstance {
	if db == nil {
		return nil
	}
	var runners []RedditRunner
	if err := db.Order("name ASC").Find(&runners).Error; err != nil {
		return nil
	}
	out := make([]p_seer_workerregistry.WorkerInstance, 0, len(runners))
	for i := range runners {
		r := runners[i]
		lastLog, err := p_seer_workerregistry.LatestWorkerRunLog(db, p_seer_workerregistry.WorkerRunnerKindReddit, r.ID)
		if err != nil {
			slog.Error("p_seer_reddit: active workers latest run log", "error", err, "runner_id", r.ID)
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
		out = append(out, &redditWorkerInstance{
			name:     r.Name,
			interval: r.Duration,
			lastRun:  lastRun,
			nextRun:  nextRun,
		})
	}
	return out
}

func init() {
	if err := p_seer_workerregistry.RegistryActiveWorkersProvider.Register("Reddit", redditActiveWorkersProvider{}); err != nil {
		log.Panic(err)
	}
}
