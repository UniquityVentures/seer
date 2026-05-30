package p_seer_reddit

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/views"
)

// redditRunnerWorkerPoolPOSTOnlyLayer rejects non-POST for worker-pool action routes.
type redditRunnerWorkerPoolPOSTOnlyLayer struct{}

func (redditRunnerWorkerPoolPOSTOnlyLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func redditRunnerWorkerPoolStartHandler(_ *views.View) http.Handler {
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
		runnerID := uint(id64)
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_reddit: worker pool start missing db", "error", dberr, "runner_id", runnerID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		ScheduleRedditRunnerWorkerPoolStart(db, runnerID)
		detailURL, err := lamu.RoutePath("seer_reddit.RedditRunnerDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("p_seer_reddit: worker pool start detail URL", "error", err, "runner_id", runnerID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

func redditRunnerWorkerPoolStopHandler(_ *views.View) http.Handler {
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
		StopRedditRunnerWorkerPool(uint(id64))
		detailURL, err := lamu.RoutePath("seer_reddit.RedditRunnerDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("p_seer_reddit: worker pool stop detail URL", "error", err, "runner_id", idStr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

func registerRedditRunnerWorkerPoolViews() {
	registerPluginView("seer_reddit.RedditRunnerWorkerPoolStartView",
		lamu.GetPageView("seer_reddit.RedditRunnerDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.worker_pool_start_post", redditRunnerWorkerPoolPOSTOnlyLayer{}).
			WithLayer("seer_reddit.reddit_runner.worker_pool_start", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: redditRunnerWorkerPoolStartHandler,
			}))

	registerPluginView("seer_reddit.RedditRunnerWorkerPoolStopView",
		lamu.GetPageView("seer_reddit.RedditRunnerDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.worker_pool_stop_post", redditRunnerWorkerPoolPOSTOnlyLayer{}).
			WithLayer("seer_reddit.reddit_runner.worker_pool_stop", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: redditRunnerWorkerPoolStopHandler,
			}))
}
