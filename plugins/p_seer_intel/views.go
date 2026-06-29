package p_seer_intel

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/views"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
	gorm_clause "gorm.io/gorm/clause"
)

type intelQueryPatcher struct{}

func (p intelQueryPatcher) Patch(v views.View, r *http.Request, db gorm.ChainInterface[Intel]) gorm.ChainInterface[Intel] {
	if start := r.URL.Query().Get("create_date_start"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			db = db.Where("created_at >= ?", t)
		}
	}
	if end := r.URL.Query().Get("create_date_end"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			db = db.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	hasEmbeddingSearch := false
	if embText := strings.TrimSpace(r.URL.Query().Get("embedding_search")); embText != "" {
		values, err := EmbedQueryText(r.Context(), embText)
		if err == nil {
			vec := pgvector.NewVector(values)
			db = db.Where("embedding IS NOT NULL").Order(gorm_clause.Expr{SQL: "embedding <=> ? ASC", Vars: []any{vec}})
			hasEmbeddingSearch = true
		} else {
			slog.Error("p_seer_intel: embed search text failed", "error", err)
		}
	}

	if !hasEmbeddingSearch {
		db = db.Order("datetime DESC, id DESC")
	}

	return db
}

func init() {
	intelListPatchers := views.QueryPatchers[Intel]{
		{Key: "seer_intel.intel.custom_patcher", Value: intelQueryPatcher{}},
	}

	registerPluginView("seer_intel.ListView",
		lamu.GetPageView("seer_intel.IntelTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_intel.intel.list", views.LayerList[Intel]{
				Key:           getters.Static("intels"),
				QueryPatchers: intelListPatchers,
			}))

	registerPluginView("seer_intel.DetailView",
		lamu.GetPageView("seer_intel.IntelDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_intel.intel.detail", views.LayerDetail[Intel]{
				Key:          getters.Static("intel"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_intel.intel.source_href", intelSourceDetailHrefLayer{}))
}
