package p_seer_gdelt

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/views"
)

// gdeltWorkerEnrichSourceIDsLayer sets [GDELTWorker.GDELTSourceIDs] on GET for the worker edit form
// so [components.InputManyToMany] shows assigned sources as chips.
type gdeltWorkerEnrichSourceIDsLayer struct{}

func (gdeltWorkerEnrichSourceIDsLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		wk, ok := ctx.Value("gdeltWorker").(GDELTWorker)
		if !ok || wk.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			next.ServeHTTP(w, r)
			return
		}
		var ids []uint
		if err := db.Model(&GDELTSource{}).Where("gdelt_worker_id = ?", wk.ID).Order("id DESC").Pluck("id", &ids).Error; err != nil {
			slog.Error("p_seer_gdelt: load worker gdelt source ids", "error", err, "worker_id", wk.ID)
			next.ServeHTTP(w, r)
			return
		}
		wk.GDELTSourceIDs = ids
		ctx = context.WithValue(ctx, "gdeltWorker", wk)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
