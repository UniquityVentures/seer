package p_seer_workerregistry

import (
	"errors"
	"time"

	"github.com/UniquityVentures/lago/lago"
	"gorm.io/gorm"
)

// Worker run kinds (discriminator for foreign runner tables).
const (
	WorkerRunnerKindReddit  = "reddit"
	WorkerRunnerKindWebsite = "website"
)

// WorkerRunLogsTable is the GORM/Postgres table for one worker-pool pass.
const WorkerRunLogsTable = "seer_worker_run_logs"

// WorkerRunStatus is the lifecycle of a single run row.
type WorkerRunStatus string

const (
	WorkerRunStatusPending WorkerRunStatus = "pending"
	WorkerRunStatusSuccess WorkerRunStatus = "success"
	WorkerRunStatusError   WorkerRunStatus = "error"
)

// WorkerRunLog records one execution of a runner worker pool (pending → success|error).
type WorkerRunLog struct {
	gorm.Model

	RunnerKind string `gorm:"size:32;not null;index:idx_wrl_lookup,priority:1"`
	RunnerID   uint   `gorm:"not null;index:idx_wrl_lookup,priority:2"`
	RunnerName string `gorm:"size:128;not null;default:''"`

	Status WorkerRunStatus `gorm:"size:16;not null;default:'pending';index"`

	StartedAt  time.Time  `gorm:"not null"`
	FinishedAt *time.Time `gorm:"index"`

	DurationMS int64 `gorm:"not null;default:0"`

	ErrorMessage string `gorm:"type:text;not null;default:''"`
}

func (WorkerRunLog) TableName() string {
	return WorkerRunLogsTable
}

func init() {
	lago.OnDBInit("p_seer_workerregistry.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[WorkerRunLog](db)
		return db
	})
}

// StartWorkerRunLog inserts a pending row at the start of a worker pass.
func StartWorkerRunLog(db *gorm.DB, kind string, runnerID uint, runnerName string) (*WorkerRunLog, error) {
	if db == nil {
		return nil, errors.New("p_seer_workerregistry: StartWorkerRunLog: nil db")
	}
	row := &WorkerRunLog{
		RunnerKind: kind,
		RunnerID:   runnerID,
		RunnerName: runnerName,
		Status:     WorkerRunStatusPending,
		StartedAt:  time.Now().UTC(),
	}
	if err := db.Create(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

// FinishWorkerRunLog marks the row success or error, sets finished time and duration.
func FinishWorkerRunLog(db *gorm.DB, log *WorkerRunLog, runErr error) error {
	if db == nil || log == nil || log.ID == 0 {
		return nil
	}
	now := time.Now().UTC()
	st := WorkerRunStatusSuccess
	errMsg := ""
	if runErr != nil {
		st = WorkerRunStatusError
		errMsg = runErr.Error()
	}
	dur := now.Sub(log.StartedAt).Milliseconds()
	return db.Model(log).Select("Status", "FinishedAt", "DurationMS", "ErrorMessage").Updates(WorkerRunLog{
		Status:       st,
		FinishedAt:   &now,
		DurationMS:   dur,
		ErrorMessage: errMsg,
	}).Error
}

// LatestWorkerRunLog returns the most recent finished run (success or error) for that runner, or nil if none.
func LatestWorkerRunLog(db *gorm.DB, kind string, runnerID uint) (*WorkerRunLog, error) {
	if db == nil {
		return nil, errors.New("p_seer_workerregistry: LatestWorkerRunLog: nil db")
	}
	var row WorkerRunLog
	err := db.Where("runner_kind = ? AND runner_id = ? AND finished_at IS NOT NULL", kind, runnerID).
		Order("finished_at DESC, id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
