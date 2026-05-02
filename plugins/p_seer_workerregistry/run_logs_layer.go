package p_seer_workerregistry

import (
	"context"
	"net/http"
	"reflect"

	"github.com/UniquityVentures/lago/components"
	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/views"
)

func structRunnerID(v any) uint {
	if v == nil {
		return 0
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return 0
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return 0
	}
	f := rv.FieldByName("ID")
	if !f.IsValid() {
		return 0
	}
	switch f.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uint(f.Uint())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if f.Int() < 0 {
			return 0
		}
		return uint(f.Int())
	default:
		return 0
	}
}

// RunnerRunLogsLayer loads [ListWorkerRunLogs] into context key [workerRunLogsDataKey] as
// [components.ObjectList][WorkerRunLog] for detail pages that embed [WorkerRunLogsBlock].
type RunnerRunLogsLayer struct {
	RunnerContextKey string
	Kind             string
}

func (l RunnerRunLogsLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ol := components.ObjectList[WorkerRunLog]{}
		id := structRunnerID(ctx.Value(l.RunnerContextKey))
		if id != 0 && l.Kind != "" {
			if db, err := getters.DBFromContext(ctx); err == nil {
				if items, err := ListWorkerRunLogs(db, l.Kind, id, 100); err == nil && items != nil {
					ol.Items = items
					ol.Total = uint64(len(items))
					ol.Number = 1
					ol.NumPages = 1
				}
			}
		}
		ctx = context.WithValue(ctx, workerRunLogsDataKey, ol)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
