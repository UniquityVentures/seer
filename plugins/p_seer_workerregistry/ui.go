package p_seer_workerregistry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/registry"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
	"gorm.io/gorm"
)

func init() {
	registerRoutes()
	registerPages()
	registerViews()
}

func registerRoutes() {
	registerPluginRoute("seer_workerregistry.DefaultRoute", lamu.Route{
		Path:    AppUrl,
		Handler: lamu.NewDynamicView("seer_workerregistry.WorkersView"),
	})
	registerPluginRoute("seer_workerregistry.ExportRoute", lamu.Route{
		Path:    AppUrl + "export/",
		Handler: p_users.RequireAuth(exportWorkersHandler{}),
	})
	registerPluginRoute("seer_workerregistry.ImportRoute", lamu.Route{
		Path:    AppUrl + "import/",
		Handler: p_users.RequireAuth(importWorkersHandler{}),
	})
}

func registerViews() {
	registerPluginView("seer_workerregistry.WorkersView",
		lamu.GetPageView("seer_workerregistry.WorkersPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}))
}

func registerPages() {
	registerPluginPage("seer_workerregistry.AppMenu", &components.SidebarMenu{
		Title: getters.Static("Worker Registry"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lamu.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Workers"),
				Url:   lamu.RoutePath("seer_workerregistry.DefaultRoute", nil),
			},
		},
	})

	registerPluginPage("seer_workerregistry.WorkersPage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lamu.DynamicPage{Name: "seer_workerregistry.AppMenu"},
		},
		Children: []components.PageInterface{
			&workerRegistryPage{Page: components.Page{Key: "seer_workerregistry.WorkersPageBody"}},
		},
	})
}

type workerRegistryPage struct {
	components.Page
}

func (p *workerRegistryPage) GetKey() string     { return p.Key }
func (p *workerRegistryPage) GetRoles() []string { return p.Roles }

func (p *workerRegistryPage) Build(ctx context.Context) Node {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return Div(Class("alert alert-error"), Text(err.Error()))
	}

	getMap, _ := ctx.Value("$get").(map[string]any)
	selectedType, _ := getMap["type"].(string)

	pairs := RegistryActiveWorkersProvider.AllStable(registry.AlphabeticalByKey[ActiveWorkersProvider]{})
	if pairs == nil || len(*pairs) == 0 {
		return Div(
			Class("container max-w-5xl mx-auto p-4 md:p-6 space-y-6"),
			Div(
				Class("flex flex-col gap-1"),
				H1(Class("text-2xl font-extrabold tracking-tight"), Text("Worker Registry")),
				P(Class("text-sm text-base-content/70"), Text("Monitor and manage scheduled background workers across the system.")),
			),
			Div(Class("text-center py-12 text-base-content/60"), Text("No worker types registered.")),
		)
	}

	if selectedType == "" && len(*pairs) > 0 {
		selectedType = (*pairs)[0].Key
	}

	dropdownOptions := []Node{}
	for _, pair := range *pairs {
		dropdownOptions = append(dropdownOptions, Option(
			Value(pair.Key),
			If(pair.Key == selectedType, Attr("selected", "")),
			Text(pair.Key),
		))
	}

	var provider ActiveWorkersProvider
	for _, pair := range *pairs {
		if pair.Key == selectedType {
			provider = pair.Value
			break
		}
	}

	var workers []WorkerInstance
	if provider != nil {
		workers = provider.FetchActiveWorkers(db)
	}

	dropdown := Select(
		Name("type"),
		Class("select select-bordered w-full max-w-xs"),
		Attr("hx-get", "/seer-workerregistry/"),
		Attr("hx-select", "#worker-entries-card"),
		Attr("hx-target", "#worker-entries-card"),
		Attr("hx-swap", "outerHTML"),
		Attr("hx-push-url", "true"),
		Attr("hx-trigger", "change"),
		Group(dropdownOptions),
	)

	return Div(
		Class("container max-w-5xl mx-auto p-4 md:p-6 space-y-6"),
		Div(
			Class("flex flex-col md:flex-row md:items-center justify-between gap-4"),
			Div(
				Class("flex flex-col gap-1"),
				H1(Class("text-2xl font-extrabold tracking-tight"), Text("Worker Registry")),
				P(Class("text-sm text-base-content/70"), Text("Monitor and manage scheduled background workers across the system.")),
			),
			Div(
				Class("flex flex-wrap items-center gap-3 shrink-0"),
				A(
					Href("/seer-workerregistry/export/"),
					Class("btn btn-outline btn-sm md:btn-md gap-2"),
					Attr("data-hx-boost", "false"),
					components.Render(components.Icon{Name: "arrow-down-tray", Classes: "heroicon-sm"}, ctx),
					Text("Export JSON"),
				),
				FormEl(
					Class("flex items-center gap-2"),
					Attr("hx-post", "/seer-workerregistry/import/"),
					Attr("hx-encoding", "multipart/form-data"),
					Attr("hx-target", "#worker-entries-card"),
					Attr("hx-swap", "outerHTML"),
					Input(
						Type("file"),
						Name("file"),
						Accept(".json"),
						Class("file-input file-input-bordered file-input-sm w-48 md:w-64"),
						Attr("onchange", "this.form.requestSubmit()"),
					),
				),
			),
		),
		Div(
			ID("worker-entries-card"),
			Class("space-y-4"),
			Div(
				Class("flex flex-col gap-2"),
				Label(Class("text-xs font-bold uppercase tracking-wider text-base-content/60"), Text("Worker Type")),
				dropdown,
			),
			renderWorkerList(ctx, workers),
		),
	)
}

func renderWorkerList(ctx context.Context, workers []WorkerInstance) Node {
	if len(workers) == 0 {
		return Div(Class("text-center py-8 text-base-content/60 text-sm border border-base-300 rounded-box bg-base-100 shadow-sm"), Text("No workers registered for this type."))
	}

	workerNodes := []Node{}
	for _, w := range workers {
		lastRunStr := "—"
		if lr := w.LastRun(); lr != nil && !lr.IsZero() {
			lastRunStr = lr.UTC().Format(time.RFC3339)
		}
		nextRunStr := "—"
		if nr := w.NextRun(); nr != nil && !nr.IsZero() {
			nextRunStr = nr.UTC().Format(time.RFC3339)
		}

		detailURL := w.DetailURL(ctx)

		workerNodes = append(workerNodes, Div(
			Class("flex flex-col md:flex-row md:items-center justify-between p-4 hover:bg-base-50/50 dark:hover:bg-base-900/10 transition-colors gap-4"),
			Div(
				Class("flex-1 min-w-0"),
				If(detailURL != "", A(Href(detailURL), Class("link link-primary hover:underline text-sm font-semibold break-all"), Text(w.Name()))),
				If(detailURL == "", Div(Class("text-sm font-semibold text-primary break-all"), Text(w.Name()))),
				Div(Class("text-xs opacity-75 mt-0.5"), Text(fmt.Sprintf("Interval: %v", w.Interval()))),
			),
			Div(
				Class("flex items-center gap-6 text-xs text-base-content/80 whitespace-nowrap"),
				Div(
					Class("flex flex-col items-start"),
					Span(Class("opacity-60 text-[10px] uppercase font-bold tracking-wider mb-0.5"), Text("Last Run")),
					Span(Class("font-mono"), Text(lastRunStr)),
				),
				Div(
					Class("flex flex-col items-start"),
					Span(Class("opacity-60 text-[10px] uppercase font-bold tracking-wider mb-0.5"), Text("Next Run")),
					Span(Class("font-mono"), Text(nextRunStr)),
				),
			),
		))
	}

	return Div(
		Class("border border-base-300 rounded-box overflow-hidden bg-base-100 divide-y divide-base-200 shadow-sm"),
		Group(workerNodes),
	)
}

type exportWorkersHandler struct{}

func (exportWorkersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type exportedWorker struct {
		Type     string `json:"type"`
		ID       uint   `json:"id"`
		Name     string `json:"name"`
		Duration int64  `json:"duration_ns"`
		Args     any    `json:"args,omitempty"`
	}

	var list []exportedWorker

	// Reddit
	var redditRunners []struct {
		ID       uint
		Name     string
		Duration int64
	}
	if err := db.Table("seer_reddit_runners").Where("deleted_at IS NULL").Scan(&redditRunners).Error; err == nil {
		for _, r := range redditRunners {
			var sources []map[string]any
			_ = db.Table("seer_reddit_sources").Where("reddit_runner_id = ? AND deleted_at IS NULL", r.ID).Scan(&sources)
			for i := range sources {
				delete(sources[i], "deleted_at")
				delete(sources[i], "reddit_runner_id")
				parseJSONFields(sources[i])
			}
			list = append(list, exportedWorker{
				Type:     "Reddit",
				ID:       r.ID,
				Name:     r.Name,
				Duration: r.Duration,
				Args:     sources,
			})
		}
	}

	// Website
	var websiteRunners []struct {
		ID       uint
		Name     string
		Duration int64
	}
	if err := db.Table("seer_website_runners").Where("deleted_at IS NULL").Scan(&websiteRunners).Error; err == nil {
		for _, r := range websiteRunners {
			var sources []map[string]any
			_ = db.Table("seer_website_sources").Where("website_runner_id = ? AND deleted_at IS NULL", r.ID).Scan(&sources)
			for i := range sources {
				delete(sources[i], "deleted_at")
				delete(sources[i], "website_runner_id")
				parseJSONFields(sources[i])
			}
			list = append(list, exportedWorker{
				Type:     "Website",
				ID:       r.ID,
				Name:     r.Name,
				Duration: r.Duration,
				Args:     sources,
			})
		}
	}

	// GDELT
	var gdeltWorkers []struct {
		ID       uint
		Name     string
		Duration int64
	}
	if err := db.Table("seer_gdelt_workers").Where("deleted_at IS NULL").Scan(&gdeltWorkers).Error; err == nil {
		for _, r := range gdeltWorkers {
			var sources []map[string]any
			_ = db.Table("seer_gdelt_sources").Where("gdelt_worker_id = ? AND deleted_at IS NULL", r.ID).Scan(&sources)
			for i := range sources {
				delete(sources[i], "deleted_at")
				delete(sources[i], "gdelt_worker_id")
				parseJSONFields(sources[i])
			}
			list = append(list, exportedWorker{
				Type:     "GDELT",
				ID:       r.ID,
				Name:     r.Name,
				Duration: r.Duration,
				Args:     sources,
			})
		}
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="workers.json"`)
	_, _ = w.Write(data)
}

func parseJSONFields(m map[string]any) {
	for k, v := range m {
		switch val := v.(type) {
		case []byte:
			var decoded any
			if json.Unmarshal(val, &decoded) == nil {
				m[k] = decoded
			}
		case string:
			trimmed := strings.TrimSpace(val)
			if (strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) ||
				(strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) {
				var decoded any
				if json.Unmarshal([]byte(trimmed), &decoded) == nil {
					m[k] = decoded
				}
			}
		}
	}
}

type importWorkersHandler struct{}

func (importWorkersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Max 10 MB file
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file parameter", http.StatusBadRequest)
		return
	}
	defer file.Close()

	var list []struct {
		Type       string           `json:"type"`
		Name       string           `json:"name"`
		DurationNS int64            `json:"duration_ns"`
		Args       []map[string]any `json:"args"`
	}

	if err := json.NewDecoder(file).Decode(&list); err != nil {
		http.Error(w, "failed to decode json: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Transaction to ensure atomic import
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, item := range list {
			var tableNameRunners string
			var tableNameSources string
			var fkName string

			switch item.Type {
			case "Reddit":
				tableNameRunners = "seer_reddit_runners"
				tableNameSources = "seer_reddit_sources"
				fkName = "reddit_runner_id"
			case "Website":
				tableNameRunners = "seer_website_runners"
				tableNameSources = "seer_website_sources"
				fkName = "website_runner_id"
			case "GDELT":
				tableNameRunners = "seer_gdelt_workers"
				tableNameSources = "seer_gdelt_sources"
				fkName = "gdelt_worker_id"
			default:
				continue
			}

			// Find or create runner
			var runnerID uint
			var existing struct {
				ID uint
			}
			err := tx.Table(tableNameRunners).Where("name = ? AND deleted_at IS NULL", item.Name).First(&existing).Error
			if err == nil {
				// Exists, update duration
				runnerID = existing.ID
				if err := tx.Table(tableNameRunners).Where("id = ?", runnerID).Update("duration", item.DurationNS).Error; err != nil {
					return err
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "record not found") {
				// Create new runner
				newRunner := map[string]any{
					"created_at": time.Now(),
					"updated_at": time.Now(),
					"name":       item.Name,
					"duration":   item.DurationNS,
				}
				if err := tx.Table(tableNameRunners).Create(&newRunner).Error; err != nil {
					return err
				}
				// Query back the ID
				if err := tx.Table(tableNameRunners).Where("name = ? AND deleted_at IS NULL", item.Name).First(&existing).Error; err != nil {
					return err
				}
				runnerID = existing.ID
			} else {
				return err
			}

			// Delete existing sources for this runner/worker
			if err := tx.Table(tableNameSources).Where(fmt.Sprintf("%s = ?", fkName), runnerID).Delete(map[string]any{}).Error; err != nil {
				return err
			}

			// Insert new sources
			for _, arg := range item.Args {
				prepared := prepareSourceMap(arg, fkName, runnerID)
				if err := tx.Table(tableNameSources).Create(&prepared).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		http.Error(w, "database transaction failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Re-render the active worker panel `#worker-entries-card`
	selectedType := r.URL.Query().Get("type")

	pairs := RegistryActiveWorkersProvider.AllStable(registry.AlphabeticalByKey[ActiveWorkersProvider]{})
	if pairs == nil || len(*pairs) == 0 {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<div id="worker-entries-card" class="text-center py-12 text-base-content/60">No worker types registered.</div>`))
		return
	}

	if selectedType == "" && len(*pairs) > 0 {
		selectedType = (*pairs)[0].Key
	}

	var provider ActiveWorkersProvider
	for _, pair := range *pairs {
		if pair.Key == selectedType {
			provider = pair.Value
			break
		}
	}

	var workers []WorkerInstance
	if provider != nil {
		workers = provider.FetchActiveWorkers(db)
	}

	dropdownOptions := []Node{}
	for _, pair := range *pairs {
		dropdownOptions = append(dropdownOptions, Option(
			Value(pair.Key),
			If(pair.Key == selectedType, Attr("selected", "")),
			Text(pair.Key),
		))
	}

	dropdown := Select(
		Name("type"),
		Class("select select-bordered w-full max-w-xs"),
		Attr("hx-get", "/seer-workerregistry/"),
		Attr("hx-select", "#worker-entries-card"),
		Attr("hx-target", "#worker-entries-card"),
		Attr("hx-swap", "outerHTML"),
		Attr("hx-push-url", "true"),
		Attr("hx-trigger", "change"),
		Group(dropdownOptions),
	)

	node := Div(
		ID("worker-entries-card"),
		Class("space-y-4"),
		Div(
			Class("flex flex-col gap-2"),
			Label(Class("text-xs font-bold uppercase tracking-wider text-base-content/60"), Text("Worker Type")),
			dropdown,
		),
		renderWorkerList(ctx, workers),
	)

	w.Header().Set("Content-Type", "text/html")
	_ = node.Render(w)
}

func prepareSourceMap(m map[string]any, fkName string, fkVal uint) map[string]any {
	out := make(map[string]any)
	for k, v := range m {
		// Skip fields that shouldn't be overridden or are auto-generated/deleted
		if k == "id" || k == "created_at" || k == "updated_at" || k == "deleted_at" || k == fkName {
			continue
		}

		// Convert slice or map values to JSON strings so GORM can save them to JSON columns
		switch val := v.(type) {
		case []any, map[string]any:
			if bytes, err := json.Marshal(val); err == nil {
				out[k] = string(bytes)
			} else {
				out[k] = val
			}
		case string:
			if (k == "start_date" || k == "end_date") && val != "" {
				if t, err := time.Parse(time.RFC3339, val); err == nil {
					out[k] = t
				} else {
					out[k] = val
				}
			} else {
				out[k] = val
			}
		default:
			out[k] = val
		}
	}
	out["created_at"] = time.Now()
	out["updated_at"] = time.Now()
	out[fkName] = fkVal
	return out
}


