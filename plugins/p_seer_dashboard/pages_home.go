package p_seer_dashboard

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/UniquityVentures/lago/components"
	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/lago/plugins/p_users"
	"github.com/UniquityVentures/lago/registry"
	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// seerDashboardWorkerTabs renders [components.ClientTabs] from [RegistryActiveWorkersProvider].
type seerDashboardWorkerTabs struct {
	components.Page
}

func (e seerDashboardWorkerTabs) GetKey() string     { return e.Key }
func (e seerDashboardWorkerTabs) GetRoles() []string { return e.Roles }

func (e seerDashboardWorkerTabs) Build(ctx context.Context) Node {
	pairs := p_seer_workerregistry.RegistryActiveWorkersProvider.AllStable(registry.AlphabeticalByKey[p_seer_workerregistry.ActiveWorkersProvider]{})
	if pairs == nil || len(*pairs) == 0 {
		return Div(Class("text-sm opacity-70 p-2"), Text("No worker providers registered."))
	}
	tabs := make(map[string]getters.Getter[components.PageInterface])
	for _, pair := range *pairs {
		k := pair.Key
		prov := pair.Value
		tabs[k] = func(ctx context.Context) (components.PageInterface, error) {
			db, err := getters.DBFromContext(ctx)
			if err != nil {
				return nil, err
			}
			return seerDashboardWorkerTabPage(k, prov.FetchActiveWorkers(db)), nil
		}
	}
	// Vertical tab ribbon (stacked labels) above worker panel content — see [components.ClientTabsLayoutVertical].
	return components.ClientTabs{
		Page:     components.Page{Key: e.Key + ".clientTabs"},
		Tabs:     tabs,
		StateKey: "seerDashWorkerTab",
		Layout:   components.ClientTabsLayoutVertical,
	}.Build(ctx)
}

func seerDashboardWorkerTabPage(tabKey string, workers []p_seer_workerregistry.WorkerInstance) components.PageInterface {
	if len(workers) == 0 {
		return &components.ContainerColumn{
			Page: components.Page{Key: "seer_dashboard.workers." + tabKey + ".empty"},
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Static("No workers.")},
			},
		}
	}
	children := make([]components.PageInterface, 0, len(workers))
	for i, w := range workers {
		w := w
		lastStr := "—"
		if t := w.LastRun(); t != nil {
			lastStr = t.UTC().Format(time.RFC3339)
		}
		nextStr := "—"
		if t := w.NextRun(); t != nil {
			nextStr = t.UTC().Format(time.RFC3339)
		}
		iv := w.Interval()
		children = append(children, &components.ContainerColumn{
			Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d", tabKey, i)},
			Classes: "rounded-box border border-base-300 p-3 gap-1 mb-2",
			Children: []components.PageInterface{
				&components.FieldTitle{Getter: getters.Static(w.Name())},
				&components.LabelInline{
					Title: "Interval",
					Children: []components.PageInterface{
						&components.FieldDuration{Getter: getters.Ref(getters.Static(iv))},
					},
				},
				&components.LabelInline{
					Title: "Last run",
					Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Static(lastStr)},
					},
				},
				&components.LabelInline{
					Title: "Next run",
					Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Static(nextStr)},
					},
				},
			},
		})
	}
	return &components.ContainerColumn{
		Page:     components.Page{Key: "seer_dashboard.workers." + tabKey},
		Classes:  "gap-2 w-full min-w-0",
		Children: children,
	}
}

// seerDashboardCategorizedApps renders [RegistryPluginCategory] with role filtering (same rules as [components.AppsGrid]).
type seerDashboardCategorizedApps struct {
	components.Page
}

func (e seerDashboardCategorizedApps) GetKey() string     { return e.Key }
func (e seerDashboardCategorizedApps) GetRoles() []string { return e.Roles }

func (e seerDashboardCategorizedApps) Build(ctx context.Context) Node {
	roleName := p_users.RoleFromContext(ctx, e.Key)
	pairs := p_seer_workerregistry.RegistryPluginCategory.AllStable(registry.AlphabeticalByKey[[]lago.Plugin]{})
	if pairs == nil || len(*pairs) == 0 {
		return Div(Class("text-sm opacity-70"), Text("No categorized apps."))
	}
	sections := Group{}
	for _, pair := range *pairs {
		category := pair.Key
		apps := pair.Value
		var filtered []lago.Plugin
		for _, app := range apps {
			if app.Type != lago.PluginTypeApp {
				continue
			}
			if roleName != "superuser" && len(app.Roles) > 0 {
				if !slices.Contains(app.Roles, roleName) {
					continue
				}
			}
			filtered = append(filtered, app)
		}
		if len(filtered) == 0 {
			continue
		}
		grid := Group{}
		for _, app := range filtered {
			grid = append(grid, A(
				Href(app.URL.String()),
				Class("btn btn-md h-auto flex-col space-y-1 py-4"),
				Attr("x-show", fmt.Sprintf("'%s'.toLowerCase().includes(search.toLowerCase())", app.VerboseName)),
				Attr("x-cloak"),
				components.Render(components.Icon{Name: app.Icon, Classes: "w-8 h-8"}, ctx),
				Div(Class("text-sm truncate min-w-0 w-full"), Text(app.VerboseName)),
			))
		}
		sections = append(sections,
			Div(Class("mb-6"),
				H3(Class("text-lg font-semibold mb-2"), Text(category)),
				Div(Class("grid grid-cols-2 @md:grid-cols-3 gap-2"), grid),
			),
		)
	}
	return Div(Class("w-full @container"), Attr("x-data", "{ search: '' }"),
		Div(Class("mb-4"),
			Input(Type("text"), Attr("x-model", "search"), Placeholder("Search apps..."), Class("input input-bordered w-full")),
		),
		sections,
	)
}

// seerDashboardIntelFeed shows recent intel from context [seerDashboardIntelLatest].
type seerDashboardIntelFeed struct {
	components.Page
}

func (e seerDashboardIntelFeed) GetKey() string     { return e.Key }
func (e seerDashboardIntelFeed) GetRoles() []string { return e.Roles }

func (e seerDashboardIntelFeed) Build(ctx context.Context) Node {
	list, err := getters.Key[components.ObjectList[p_seer_intel.Intel]]("seerDashboardIntelLatest")(ctx)
	if err != nil || len(list.Items) == 0 {
		return Div(Class("text-sm opacity-70"), Text("No intel yet."))
	}
	nodes := Group{}
	for _, it := range list.Items {
		it := it
		title := strings.TrimSpace(it.Title)
		if title == "" {
			title = "Untitled"
		}
		summary := strings.TrimSpace(it.Summary)
		if len(summary) > 280 {
			summary = summary[:277] + "…"
		}
		href, herr := lago.RoutePath("seer_intel.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(it.ID)),
		})(ctx)
		if herr != nil {
			href = "#"
		}
		dt := it.Datetime.UTC().Format(time.RFC3339)
		nodes = append(nodes, Div(Class("rounded-box border border-base-300 p-3 mb-2 space-y-1"),
			A(Href(href), Class("link link-hover font-medium text-sm"), Text(title)),
			Div(Class("text-xs opacity-70"), Text(dt)),
			P(Class("text-sm opacity-90 line-clamp-4"), Text(summary)),
		))
	}
	return Div(Class("flex flex-col gap-1 min-w-0"), nodes)
}

func registerSeerDashboardHomePagePatch() {
	lago.RegistryPage.Patch("dashboard.AppsPage", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellTopbarScaffold)
		if !ok {
			log.Panic("dashboard.AppsPage was not *components.ShellTopbarScaffold")
		}
		scaffold.ExtraHead = append(scaffold.ExtraHead, &components.MapDisplayLibreHead{
			Page: components.Page{Key: "seer_dashboard.MapLibreHead"},
		})
		components.ReplaceChild(scaffold, "dashboard.AppsPageLayout", func(layout *components.LayoutSimple) *components.LayoutSimple {
			layout.Children = []components.PageInterface{
				&components.ContainerRow{
					Page:    components.Page{Key: "seer_dashboard.HomeRow"},
					Classes: "flex flex-col xl:flex-row gap-4 w-full max-w-[1600px] mx-auto items-start",
					Children: []components.PageInterface{
						&components.ContainerColumn{
							Page:    components.Page{Key: "seer_dashboard.LeftCol"},
							Classes: "w-full xl:w-72 shrink-0",
							Children: []components.PageInterface{
								&components.FieldTitle{Getter: getters.Static("Workers")},
								&seerDashboardWorkerTabs{Page: components.Page{Key: "seer_dashboard.WorkerTabs"}},
							},
						},
						&components.ContainerColumn{
							Page:    components.Page{Key: "seer_dashboard.CenterCol"},
							Classes: "flex-1 min-w-0 gap-4",
							Children: []components.PageInterface{
								&SeerDashboardMap{
									Page:    components.Page{Key: "seer_dashboard.DashboardMap"},
									DataURL: lago.RoutePath("seer_dashboard.MapDataRoute", nil),
									Classes: "w-full h-[min(48vh,420px)] min-h-64 rounded-box border border-base-300 relative z-[1]",
								},
								&components.FieldTitle{Getter: getters.Static("Apps")},
								&seerDashboardCategorizedApps{Page: components.Page{Key: "seer_dashboard.CategorizedApps"}},
							},
						},
						&components.ContainerColumn{
							Page:    components.Page{Key: "seer_dashboard.RightCol"},
							Classes: "w-full xl:w-80 shrink-0",
							Children: []components.PageInterface{
								&components.FieldTitle{Getter: getters.Static("Intel")},
								&seerDashboardIntelFeed{Page: components.Page{Key: "seer_dashboard.IntelFeed"}},
							},
						},
					},
				},
			}
			return layout
		})
		return scaffold
	})
}
