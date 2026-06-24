package p_seer_reddit

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/views"
	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
	"gorm.io/gorm"
)

// redditPostListBySourceFlagLayer sets [redditPostListBySource] so list toolbars can use [getters.Key] (vs. error-swallowing checks on [redditSource.ID]).
type redditPostListBySourceFlagLayer struct{ Value bool }

func (l redditPostListBySourceFlagLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "redditPostListBySource", l.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// redditPostIntelContextLayer fills [redditPostIntelAddVisible], [redditPostIntelLinkVisible], [redditPostIntelDetailHref] after [views.LayerDetail] for [RedditPost].
type redditPostIntelContextLayer struct{}

func (redditPostIntelContextLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		post, ok := ctx.Value("redditPost").(RedditPost)
		if !ok {
			post = RedditPost{}
		}
		setEmpty := func() {
			ctx = context.WithValue(ctx, "redditPostIntelAddVisible", false)
			ctx = context.WithValue(ctx, "redditPostIntelLinkVisible", false)
			ctx = context.WithValue(ctx, "redditPostIntelDetailHref", "")
		}
		if post.ID == 0 {
			setEmpty()
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db, err := getters.DBFromContext(ctx)
		if err != nil {
			slog.Error("seer_reddit: reddit post intel context: db", "error", err)
			setEmpty()
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		exists, err := p_seer_intel.IntelExistsForSource(ctx, db, (RedditPost{}).Kind(), post.ID)
		if err != nil {
			slog.Error("seer_reddit: reddit post intel context: exists check", "error", err)
			exists = false
		}
		href := ""
		if exists {
			href, err = p_seer_intel.IntelDetailPathForSource(ctx, (RedditPost{}).Kind(), post.ID)
			if err != nil {
				slog.Error("seer_reddit: reddit post intel context: detail path", "error", err)
				href = ""
			}
		}
		ctx = context.WithValue(ctx, "redditPostIntelAddVisible", !exists)
		ctx = context.WithValue(ctx, "redditPostIntelLinkVisible", exists)
		ctx = context.WithValue(ctx, "redditPostIntelDetailHref", href)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// redditRunnerWorkerPoolStateLayer sets [workerPoolIsRunning] after [views.LayerDetail] for [RedditRunner].
type redditRunnerWorkerPoolStateLayer struct{}

func (redditRunnerWorkerPoolStateLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		run, ok := ctx.Value("redditRunner").(RedditRunner)
		running := ok && run.ID != 0 && RedditRunnerWorkerPoolIsRunning(run.ID)
		ctx = context.WithValue(ctx, "workerPoolIsRunning", running)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func init() {
	sourcePatchers := views.QueryPatchers[RedditSource]{
		{Key: "seer_reddit.source.order", Value: views.QueryPatcherOrderBy[RedditSource]{Order: "id DESC"}},
	}
	sourceDetailPatchers := views.QueryPatchers[RedditSource]{
		{Key: "seer_reddit.source.preload_runner", Value: views.QueryPatcherPreload[RedditSource]{Fields: []string{"RedditRunner"}}},
	}
	sourceUnsetPatchers := views.QueryPatchers[RedditSource]{
		{Key: "seer_reddit.source.unset_runner", Value: redditSourceUnsetRunnerPatcher{}},
		{Key: "seer_reddit.source.order", Value: views.QueryPatcherOrderBy[RedditSource]{Order: "id DESC"}},
	}

	registerPluginView("seer_reddit.RedditSourceListView",
		lamu.GetPageView("seer_reddit.RedditSourceTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.list", views.LayerList[RedditSource]{
				Key:           getters.Static("redditSources"),
				QueryPatchers: sourcePatchers,
			}))

	registerPluginView("seer_reddit.RedditSourceUnsetSelectView",
		lamu.GetPageView("seer_reddit.RedditSourceUnsetSelectionTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.unset_select_list", views.LayerList[RedditSource]{
				Key:           getters.Static("redditSources"),
				QueryPatchers: sourceUnsetPatchers,
			}))

	registerPluginView("seer_reddit.RedditSourceDetailView",
		lamu.GetPageView("seer_reddit.RedditSourceDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.detail", views.LayerDetail[RedditSource]{
				Key:           getters.Static("redditSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: sourceDetailPatchers,
			}))

	registerPluginView("seer_reddit.RedditSourceCreateView",
		lamu.GetPageView("seer_reddit.RedditSourceCreateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.create", views.LayerCreate[RedditSource]{
				SuccessURL: lamu.RoutePath("seer_reddit.RedditSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_reddit.source.create_validate", Value: redditSourceCreateValidate{}},
				},
			}))

	registerPluginView("seer_reddit.RedditSourceUpdateView",
		lamu.GetPageView("seer_reddit.RedditSourceUpdateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.detail", views.LayerDetail[RedditSource]{
				Key:           getters.Static("redditSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: sourceDetailPatchers,
			}).
			WithLayer("seer_reddit.reddit_source.update", views.LayerUpdate[RedditSource]{
				Key: getters.Static("redditSource"),
				SuccessURL: lamu.RoutePath("seer_reddit.RedditSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditSource.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_reddit.source.create_validate", Value: redditSourceCreateValidate{}},
				},
			}))

	registerPluginView("seer_reddit.RedditSourceDeleteView",
		lamu.GetPageView("seer_reddit.RedditSourceDeleteForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_source.delete_detail", views.LayerDetail[RedditSource]{
				Key:           getters.Static("redditSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: sourceDetailPatchers,
			}).
			WithLayer("seer_reddit.reddit_source.delete", views.LayerDelete[RedditSource]{
				Key:        getters.Static("redditSource"),
				SuccessURL: lamu.RoutePath("seer_reddit.DefaultRoute", nil),
			}))

	postPatchers := views.QueryPatchers[RedditPost]{
		{Key: "seer_reddit.post.not_deleted", Value: redditPostActiveOnlyPatcher{}},
		{Key: "seer_reddit.post.order", Value: views.QueryPatcherOrderBy[RedditPost]{Order: "id DESC"}},
	}

	postDetailPatchers := views.QueryPatchers[RedditPost]{
		{Key: "seer_reddit.post_detail.not_deleted", Value: redditPostActiveOnlyPatcher{}},
	}

	registerPluginView("seer_reddit.RedditPostListView",
		lamu.GetPageView("seer_reddit.RedditPostTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.post_list_by_source_flag", redditPostListBySourceFlagLayer{Value: false}).
			WithLayer("seer_reddit.reddit_post.list", views.LayerList[RedditPost]{
				Key:           getters.Static("redditPosts"),
				QueryPatchers: postPatchers,
			}))

	registerPluginView("seer_reddit.RedditPostListBySourceView",
		lamu.GetPageView("seer_reddit.RedditPostTableBySource").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.post_list_by_source_flag", redditPostListBySourceFlagLayer{Value: true}).
			WithLayer("seer_reddit.reddit_source.detail_by_source_id", views.LayerDetail[RedditSource]{
				Key:           getters.Static("redditSource"),
				PathParamKey:  getters.Static("source_id"),
				QueryPatchers: sourceDetailPatchers,
			}).
			WithLayer("seer_reddit.reddit_post.list_by_source", views.LayerList[RedditPost]{
				Key:           getters.Static("redditPosts"),
				QueryPatchers: redditPostListQueryPatchersForSource(),
			}))

	registerPluginView("seer_reddit.RedditSourceFetchPostsView",
		lamu.GetPageView("seer_reddit.RedditPostTableBySource").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.post_list_by_source_flag", redditPostListBySourceFlagLayer{Value: true}).
			WithLayer("seer_reddit.fetch_posts", redditSourceFetchPostsActionLayer{}))

	registerPluginView("seer_reddit.RedditPostDetailView",
		lamu.GetPageView("seer_reddit.RedditPostDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_post.detail", views.LayerDetail[RedditPost]{
				Key:           getters.Static("redditPost"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: postDetailPatchers,
			}).
			WithLayer("seer_reddit.reddit_post.intel", redditPostIntelContextLayer{}))

	registerPluginView("seer_reddit.RedditPostSoftDeleteView",
		lamu.GetPageView("seer_reddit.RedditPostDeleteForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_post.delete_detail", views.LayerDetail[RedditPost]{
				Key:           getters.Static("redditPost"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: postDetailPatchers,
			}).
			WithLayer("seer_reddit.reddit_post.soft_delete", redditPostSoftDeleteLayer{}))

	runnerPatchers := views.QueryPatchers[RedditRunner]{
		{Key: "seer_reddit.runner.order", Value: views.QueryPatcherOrderBy[RedditRunner]{Order: "id DESC"}},
	}

	registerPluginView("seer_reddit.RedditRunnerListView",
		lamu.GetPageView("seer_reddit.RedditRunnerTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.list", views.LayerList[RedditRunner]{
				Key:           getters.Static("redditRunners"),
				QueryPatchers: runnerPatchers,
			}))

	registerPluginView("seer_reddit.RedditRunnerSelectView",
		lamu.GetPageView("seer_reddit.RedditRunnerSelectionTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.select_list", views.LayerList[RedditRunner]{
				Key:           getters.Static("redditRunners"),
				QueryPatchers: runnerPatchers,
			}))

	registerPluginView("seer_reddit.RedditRunnerDetailView",
		lamu.GetPageView("seer_reddit.RedditRunnerDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.detail", views.LayerDetail[RedditRunner]{
				Key:          getters.Static("redditRunner"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_reddit.reddit_runner.worker_pool_state", redditRunnerWorkerPoolStateLayer{}).
			WithLayer("seer_reddit.reddit_runner.run_logs", p_seer_workerregistry.RunnerRunLogsLayer{
				RunnerContextKey: "redditRunner",
				Kind:             p_seer_workerregistry.WorkerRunnerKindReddit,
			}))

	registerPluginView("seer_reddit.RedditRunnerCreateView",
		lamu.GetPageView("seer_reddit.RedditRunnerCreateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.create", views.LayerCreate[RedditRunner]{
				SuccessURL: lamu.RoutePath("seer_reddit.RedditRunnerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_reddit.runner.validate", Value: redditRunnerValidate{}},
				},
			}))

	registerPluginView("seer_reddit.RedditRunnerUpdateView",
		lamu.GetPageView("seer_reddit.RedditRunnerUpdateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.detail", views.LayerDetail[RedditRunner]{
				Key:          getters.Static("redditRunner"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_reddit.reddit_runner.enrich_source_ids", redditRunnerEnrichSourceIDsLayer{}).
			WithLayer("seer_reddit.reddit_runner.update", views.LayerUpdate[RedditRunner]{
				Key: getters.Static("redditRunner"),
				SuccessURL: lamu.RoutePath("seer_reddit.RedditRunnerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditRunner.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_reddit.runner.validate", Value: redditRunnerValidate{}},
				},
			}))

	registerPluginView("seer_reddit.RedditRunnerDeleteView",
		lamu.GetPageView("seer_reddit.RedditRunnerDeleteForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_reddit.reddit_runner.delete_detail", views.LayerDetail[RedditRunner]{
				Key:          getters.Static("redditRunner"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_reddit.reddit_runner.delete", views.LayerDelete[RedditRunner]{
				Key:        getters.Static("redditRunner"),
				SuccessURL: lamu.RoutePath("seer_reddit.RedditRunnerListRoute", nil),
			}))
}

func redditPostListQueryPatchersForSource() views.QueryPatchers[RedditPost] {
	return views.QueryPatchers[RedditPost]{
		{Key: "seer_reddit.post.not_deleted", Value: redditPostActiveOnlyPatcher{}},
		{Key: "seer_reddit.post.order", Value: views.QueryPatcherOrderBy[RedditPost]{Order: "id DESC"}},
		{Key: "seer_reddit.post.for_current_source", Value: redditPostsForCurrentSourcePatcher{}},
	}
}

// redditPostActiveOnlyPatcher scopes queries to rows with [gorm.Model.DeletedAt] unset (non–soft-deleted).
type redditPostActiveOnlyPatcher struct{}

func (redditPostActiveOnlyPatcher) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[RedditPost]) gorm.ChainInterface[RedditPost] {
	return q.Where("deleted_at IS NULL")
}

type redditSourceUnsetRunnerPatcher struct{}

func (redditSourceUnsetRunnerPatcher) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[RedditSource]) gorm.ChainInterface[RedditSource] {
	return q.Where("reddit_runner_id IS NULL")
}

type redditSourceCreateValidate struct{}

func (redditSourceCreateValidate) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	p, err := RedditSourceCreateParamsFromFormMap(formData)
	if err != nil {
		formErrors["Subreddits"] = err
		return formData, formErrors
	}
	for k, v := range ValidateRedditSourceCreate(p) {
		formErrors[k] = v
	}
	return formData, formErrors
}

func redditRunnerIDFromFormMap(formData map[string]any) (*uint, bool) {
	v, ok := formData["RedditRunnerID"]
	if !ok {
		return nil, false
	}
	rid, ok := v.(uint)
	if !ok {
		return nil, false
	}
	if rid == 0 {
		return nil, true
	}
	return new(rid), true
}

type redditRunnerValidate struct{}

func (redditRunnerValidate) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if formErrors == nil {
		formErrors = map[string]error{}
	}
	name, _ := formData["Name"].(string)
	if strings.TrimSpace(name) == "" {
		formErrors["Name"] = errors.New("name is required")
	}
	durRaw, ok := formData["Duration"]
	if !ok {
		formErrors["Duration"] = errors.New("duration is required")
		return formData, formErrors
	}
	d, ok := durRaw.(*time.Duration)
	if !ok {
		formErrors["Duration"] = errors.New("invalid duration")
		return formData, formErrors
	}
	if d == nil || *d <= 0 {
		formErrors["Duration"] = errors.New("duration must be positive")
	}
	formData, formErrors = redditRunnerSourceIDsValidateAndFlatten(formData, formErrors)
	formErrors = validateRedditRunnerSourceIDs(r, formData, formErrors)
	return formData, formErrors
}

func redditRunnerSourceIDsValidateAndFlatten(formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	raw, ok := formData["RedditSourceIDs"]
	if !ok {
		return formData, formErrors
	}
	assoc, ok := raw.(components.AssociationIDs)
	if !ok {
		formErrors["RedditSourceIDs"] = errors.New("invalid Reddit sources")
		delete(formData, "RedditSourceIDs")
		return formData, formErrors
	}
	formData["RedditSourceIDs"] = assoc.IDs
	return formData, formErrors
}

func validateRedditRunnerSourceIDs(r *http.Request, formData map[string]any, formErrors map[string]error) map[string]error {
	ids, _ := formData["RedditSourceIDs"].([]uint)
	if len(ids) == 0 || formErrors["RedditSourceIDs"] != nil {
		return formErrors
	}
	db, err := getters.DBFromContext(r.Context())
	if err != nil {
		formErrors["RedditSourceIDs"] = err
		return formErrors
	}
	query := db.WithContext(r.Context()).Model(&RedditSource{}).Where("id IN ?", ids)
	if runner, ok := r.Context().Value("redditRunner").(RedditRunner); ok && runner.ID != 0 {
		query = query.Where("reddit_runner_id IS NULL OR reddit_runner_id = ?", runner.ID)
	} else {
		query = query.Where("reddit_runner_id IS NULL")
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		formErrors["RedditSourceIDs"] = err
		return formErrors
	}
	if count != int64(len(ids)) {
		formErrors["RedditSourceIDs"] = errors.New("select only Reddit sources without workers")
	}
	return formErrors
}
