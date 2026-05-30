package p_seer_gdelt

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/views"
)

type gdeltWorkerPoolPOSTOnlyLayer struct{}

func (gdeltWorkerPoolPOSTOnlyLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func gdeltWorkerPoolStartHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil || id64 == 0 {
			http.NotFound(w, r)
			return
		}
		workerID := uint(id64)
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_gdelt: worker pool start missing db", "error", dberr, "worker_id", workerID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		ScheduleGDELTWorkerPoolStart(db, workerID)
		detailURL, err := lamu.RoutePath("seer_gdelt.GDELTWorkerDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("p_seer_gdelt: worker pool start detail URL", "error", err, "worker_id", workerID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

func gdeltWorkerPoolStopHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil || id64 == 0 {
			http.NotFound(w, r)
			return
		}
		StopGDELTWorkerPool(uint(id64))
		detailURL, err := lamu.RoutePath("seer_gdelt.GDELTWorkerDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("p_seer_gdelt: worker pool stop detail URL", "error", err, "worker_id", idStr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

func registerGDELTWorkerPoolViews() {
	registerPluginView("seer_gdelt.GDELTWorkerPoolStartView",
		lamu.GetPageView("seer_gdelt.GDELTWorkerDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.worker_pool_start_post", gdeltWorkerPoolPOSTOnlyLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.worker_pool_start", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: gdeltWorkerPoolStartHandler,
			}))

	registerPluginView("seer_gdelt.GDELTWorkerPoolStopView",
		lamu.GetPageView("seer_gdelt.GDELTWorkerDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.worker_pool_stop_post", gdeltWorkerPoolPOSTOnlyLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.worker_pool_stop", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: gdeltWorkerPoolStopHandler,
			}))
}
