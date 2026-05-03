package p_seer_gdelt

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/UniquityVentures/lago/syncmap"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
	"gorm.io/gorm"
)

// gdeltWorkerPoolCancels is one cooperative goroutine per [GDELTWorker]: it loads sources
// with this worker id, runs [FetchAndStoreGDELTEvents] for each, then sleeps [GDELTWorker.Duration]
// before the next pass (until stopped or worker row missing).
var gdeltWorkerPoolCancels = &syncmap.SyncMap[uint, context.CancelFunc]{}

// GDELTWorkerPoolIsRunning reports whether a worker-pool goroutine is registered for workerID.
func GDELTWorkerPoolIsRunning(workerID uint) bool {
	if workerID == 0 {
		return false
	}
	cancel, ok := gdeltWorkerPoolCancels.Load(workerID)
	return ok && cancel != nil
}

// StopGDELTWorkerPool cancels the pool goroutine for workerID and removes it from the map.
func StopGDELTWorkerPool(workerID uint) {
	if workerID == 0 {
		return
	}
	cancel, loaded := gdeltWorkerPoolCancels.LoadAndDelete(workerID)
	if loaded && cancel != nil {
		cancel()
	}
}

// StartAllGDELTWorkerPools schedules a worker pool goroutine for every [GDELTWorker] row.
// It is invoked from [lago.OnDBInit] after models register so the process autostarts pools on boot.
func StartAllGDELTWorkerPools(db *gorm.DB) {
	if db == nil {
		return
	}
	var ids []uint
	if err := db.Model(&GDELTWorker{}).Order("id ASC").Pluck("id", &ids).Error; err != nil {
		slog.Error("p_seer_gdelt: autostart list workers", "error", err)
		return
	}
	for _, id := range ids {
		if id == 0 {
			continue
		}
		ScheduleGDELTWorkerPoolStart(db, id)
	}
	slog.Info("p_seer_gdelt: worker pools autostart scheduled", "count", len(ids))
}

// ScheduleGDELTWorkerPoolStart starts the pool in a new goroutine if not already running.
// db must be a pooled *gorm.DB (not an open transaction).
func ScheduleGDELTWorkerPoolStart(db *gorm.DB, workerID uint) {
	if db == nil || workerID == 0 {
		return
	}
	d := db
	id := workerID
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		if _, loaded := gdeltWorkerPoolCancels.LoadOrStore(id, cancel); loaded {
			cancel()
			return
		}
		runGDELTWorkerPool(d.WithContext(ctx), id, ctx)
	}()
}

func runGDELTWorkerPool(db *gorm.DB, workerID uint, ctx context.Context) {
	defer slog.Info("p_seer_gdelt: worker pool exited", "worker_id", workerID)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var worker GDELTWorker
		if err := db.First(&worker, workerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Warn("p_seer_gdelt: worker pool worker row missing", "worker_id", workerID)
			} else {
				slog.Error("p_seer_gdelt: worker pool load worker", "error", err, "worker_id", workerID)
			}
			StopGDELTWorkerPool(workerID)
			return
		}

		var runLog *p_seer_workerregistry.WorkerRunLog
		if row, err := p_seer_workerregistry.StartWorkerRunLog(db, p_seer_workerregistry.WorkerRunnerKindGDELT, workerID, worker.Name); err != nil {
			slog.Error("p_seer_gdelt: worker run log start", "error", err, "worker_id", workerID)
		} else {
			runLog = row
		}

		var sources []GDELTSource
		var runErr error
		if err := db.Where("gdelt_worker_id = ?", workerID).Find(&sources).Error; err != nil {
			runErr = err
			slog.Error("p_seer_gdelt: worker pool list sources", "error", err, "worker_id", workerID)
		} else {
			for i := range sources {
				src := sources[i]
				req := src.ToGDELTSearchRequest()
				sid := src.ID
				stored, err := FetchAndStoreGDELTEvents(ctx, db, req, &sid)
				if err != nil {
					runErr = errors.Join(runErr, err)
					slog.Error("p_seer_gdelt: worker pool fetch",
						"error", err,
						"worker_id", workerID,
						"gdelt_source_id", src.ID,
					)
				} else if len(stored) > 0 {
					rows := append([]Event(nil), stored...)
					dbCopy := db
					go RunGDELTEventsIntelIngest(context.Background(), dbCopy, rows)
				}
			}
		}
		if runLog != nil {
			if err := p_seer_workerregistry.FinishWorkerRunLog(db, runLog, runErr); err != nil {
				slog.Error("p_seer_gdelt: worker run log finish", "error", err, "worker_id", workerID)
			}
		}

		if worker.Duration <= 0 {
			StopGDELTWorkerPool(workerID)
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(worker.Duration):
		}
	}
}
