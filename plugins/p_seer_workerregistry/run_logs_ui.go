package p_seer_workerregistry

import (
	"context"
	"fmt"
	"time"

	"github.com/UniquityVentures/lago/components"
	"github.com/UniquityVentures/lago/getters"
)

// workerRunLogsDataKey is the context key [views] layers set to
// [components.ObjectList][WorkerRunLog] for detail pages.
const workerRunLogsDataKey = "workerRunLogs"

func workerRunLogFinishedCell(ctx context.Context) (string, error) {
	raw, err := getters.Key[any]("$row.FinishedAt")(ctx)
	if err != nil {
		return "", err
	}
	if raw == nil {
		return "—", nil
	}
	switch t := raw.(type) {
	case *time.Time:
		if t == nil || t.IsZero() {
			return "—", nil
		}
		loc := components.DefaultTimeZone
		if tz, ok := ctx.Value("$tz").(*time.Location); ok && tz != nil {
			loc = tz
		}
		return t.In(loc).Format("Mon, 02 Jan 2006 15:04:05"), nil
	default:
		return "—", nil
	}
}

func workerRunLogStatusCell(ctx context.Context) (string, error) {
	raw, err := getters.Key[any]("$row.Status")(ctx)
	if err != nil {
		return "", err
	}
	if raw == nil {
		return "", nil
	}
	switch v := raw.(type) {
	case string:
		return v, nil
	case WorkerRunStatus:
		return string(v), nil
	default:
		return fmt.Sprintf("%v", raw), nil
	}
}

func workerRunLogDurationCell(ctx context.Context) (string, error) {
	ms, err := getters.Key[int64]("$row.DurationMS")(ctx)
	if err != nil {
		return "", err
	}
	if ms < 1000 {
		return fmt.Sprintf("%d ms", ms), nil
	}
	return fmt.Sprintf("%.1f s", float64(ms)/1000.0), nil
}

func workerRunLogErrorCell(ctx context.Context) (string, error) {
	s, err := getters.Key[string]("$row.ErrorMessage")(ctx)
	if err != nil {
		return "", err
	}
	if s == "" {
		return "—", nil
	}
	const max = 96
	if len(s) > max {
		return s[:max-1] + "…", nil
	}
	return s, nil
}

// WorkerRunLogsBlock renders a run-history table; context must provide [workerRunLogsDataKey]
// as [components.ObjectList][WorkerRunLog]. Register [RunnerRunLogsLayer] on the detail view.
func WorkerRunLogsBlock(tableUID string) components.PageInterface {
	if tableUID == "" {
		tableUID = "worker-run-logs"
	}
	key := "seer_workerregistry.RunLogsTable." + tableUID
	return &components.LabelNewline{
		Page:  components.Page{Key: "seer_workerregistry.RunLogsSection." + tableUID},
		Title: "Run history",
		Children: []components.PageInterface{
			&components.DataTable[WorkerRunLog]{
				Page:    components.Page{Key: key},
				UID:     tableUID,
				Classes: "w-full max-w-5xl",
				Data:    getters.Key[components.ObjectList[WorkerRunLog]](workerRunLogsDataKey),
				Columns: []components.TableColumn{
					{
						Label: "Started",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.StartedAt")},
						},
					},
					{
						Label: "Finished",
						Children: []components.PageInterface{
							&components.FieldText{Getter: workerRunLogFinishedCell, Classes: "whitespace-nowrap"},
						},
					},
					{
						Label: "Duration",
						Children: []components.PageInterface{
							&components.FieldText{Getter: workerRunLogDurationCell},
						},
					},
					{
						Label: "Status",
						Children: []components.PageInterface{
							&components.FieldText{Getter: workerRunLogStatusCell},
						},
					},
					{
						Label: "Error",
						Children: []components.PageInterface{
							&components.FieldText{Getter: workerRunLogErrorCell, Classes: "whitespace-normal max-w-md"},
						},
					},
				},
			},
		},
	}
}
