package p_seer_dashboard

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/UniquityVentures/lago/components"
	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/lago/registry"
	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
	. "maragu.dev/gomponents"
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
		return (&components.FieldText{
			Page:    components.Page{Key: e.Key + ".noProviders"},
			Getter:  getters.Static("No worker providers registered."),
			Classes: "text-sm opacity-70 p-2",
		}).Build(ctx)
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
		lastRun := components.PageInterface(&components.FieldText{
			Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d.lastRun", tabKey, i)},
			Getter:  getters.Static("—"),
			Classes: "text-sm opacity-90 min-w-0",
		})
		if t := w.LastRun(); t != nil {
			tt := *t
			lastRun = &components.FieldDatetime{
				Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d.lastRun", tabKey, i)},
				Getter:  getters.Static(tt),
				Classes: "text-sm opacity-90 min-w-0",
			}
		}
		nextRun := components.PageInterface(&components.FieldText{
			Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d.nextRun", tabKey, i)},
			Getter:  getters.Static("—"),
			Classes: "text-sm opacity-90 min-w-0",
		})
		if t := w.NextRun(); t != nil {
			tt := *t
			nextRun = &components.FieldDatetime{
				Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d.nextRun", tabKey, i)},
				Getter:  getters.Static(tt),
				Classes: "text-sm opacity-90 min-w-0",
			}
		}
		iv := w.Interval()
		children = append(children, &components.ContainerColumn{
			Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d", tabKey, i)},
			Classes: "rounded-box border border-base-300 p-3 mb-2",
			Children: []components.PageInterface{
				&components.FieldText{
					Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d.name", tabKey, i)},
					Getter:  getters.Static(w.Name()),
					Classes: "text-sm font-bold text-primary",
				},
				&components.FieldText{
					Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d.interval", tabKey, i)},
					Getter:  getters.Format("Runs Every %v", getters.Any(getters.Static(iv))),
					Classes: "text-sm opacity-70",
				},
				&components.ContainerRow{
					Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d.lastRunRow", tabKey, i)},
					Classes: "items-baseline min-w-0 w-full",
					Children: []components.PageInterface{
						&components.FieldText{
							Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d.lastRunLabel", tabKey, i)},
							Getter:  getters.Static("Last run"),
							Classes: "text-sm opacity-70 shrink-0",
						},
						lastRun,
					},
				},
				&components.ContainerRow{
					Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d.nextRunRow", tabKey, i)},
					Classes: "items-baseline min-w-0 w-full",
					Children: []components.PageInterface{
						&components.FieldText{
							Page:    components.Page{Key: fmt.Sprintf("seer_dashboard.workers.%s.%d.nextRunLabel", tabKey, i)},
							Getter:  getters.Static("Next run"),
							Classes: "text-sm opacity-70 shrink-0",
						},
						nextRun,
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

// seerDashboardIntelFeed shows recent intel from context [seerDashboardIntelLatest].
type seerDashboardIntelFeed struct {
	components.Page
}

func (e seerDashboardIntelFeed) GetKey() string     { return e.Key }
func (e seerDashboardIntelFeed) GetRoles() []string { return e.Roles }

func (e seerDashboardIntelFeed) Build(ctx context.Context) Node {
	list, err := getters.Key[components.ObjectList[p_seer_intel.Intel]]("seerDashboardIntelLatest")(ctx)
	if err != nil || len(list.Items) == 0 {
		return (&components.FieldText{
			Page:    components.Page{Key: e.Key + ".empty"},
			Getter:  getters.Static("No intel yet."),
			Classes: "text-sm opacity-70",
		}).Build(ctx)
	}
	displayItems := list.Items
	if len(displayItems) > 5 {
		displayItems = displayItems[:5]
	}
	cards := make([]components.PageInterface, 0, len(displayItems))
	for i, it := range displayItems {
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
		cards = append(cards, &components.ContainerColumn{
			Page:    components.Page{Key: fmt.Sprintf("%s.%d", e.Key, i)},
			Classes: "rounded-box border border-base-300 p-3 mb-2",
			Children: []components.PageInterface{
				&components.FieldLink{
					Page:    components.Page{Key: fmt.Sprintf("%s.%d.link", e.Key, i)},
					Href:    getters.Static(href),
					Label:   getters.Static(title),
					Classes: "link link-hover font-bold text-sm",
				},
				&components.FieldDatetime{
					Page:    components.Page{Key: fmt.Sprintf("%s.%d.dt", e.Key, i)},
					Getter:  getters.Static(it.Datetime),
					Classes: "text-xs opacity-70",
				},
				&components.FieldText{
					Page:    components.Page{Key: fmt.Sprintf("%s.%d.summary", e.Key, i)},
					Getter:  getters.Static(summary),
					Classes: "text-sm opacity-90 line-clamp-4",
				},
			},
		})
	}
	pageChildren := make([]components.PageInterface, 0, len(cards)+1)
	pageChildren = append(pageChildren, cards...)
	if list.Total > 5 {
		pageChildren = append(pageChildren, &components.ButtonLink{
			Page:    components.Page{Key: e.Key + ".readMore"},
			Label:   "Read more",
			Link:    lago.RoutePath("seer_intel.DefaultRoute", nil),
			Classes: "btn-sm btn-outline w-full mt-1",
		})
	}
	return (&components.ContainerColumn{
		Page:     components.Page{Key: e.Key + ".list"},
		Classes:  "min-w-0",
		Children: pageChildren,
	}).Build(ctx)
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
			if len(layout.Children) != 1 {
				log.Panic("dashboard.AppsPageLayout: expected exactly one child (AppsGrid)")
			}
			appsGrid := layout.Children[0]
			layout.Children = []components.PageInterface{
				&components.ContainerRow{
					Page:    components.Page{Key: "seer_dashboard.HomeRow"},
					Classes: "flex flex-col xl:flex-row gap-4 w-full max-w-[1600px] mx-auto items-start",
					Children: []components.PageInterface{
						&components.ContainerColumn{
							Page:    components.Page{Key: "seer_dashboard.LeftCol"},
							Classes: "w-full xl:w-80 shrink-0",
							Children: []components.PageInterface{
								&components.FieldTitle{Getter: getters.Static("Intel")},
								&seerDashboardIntelFeed{Page: components.Page{Key: "seer_dashboard.IntelFeed"}},
							},
						},
						&components.ContainerColumn{
							Page:    components.Page{Key: "seer_dashboard.CenterCol"},
							Classes: "flex-1 min-w-0 gap-4",
							Children: []components.PageInterface{
								&SeerDashboardMap{
									Page:    components.Page{Key: "seer_dashboard.DashboardMap"},
									DataURL: lago.RoutePath("seer_dashboard.MapDataRoute", nil),
									Classes: "w-full h-[50vh] min-h-64 rounded-box border border-base-300 relative z-[1]",
								},
								appsGrid,
							},
						},
						&components.ContainerColumn{
							Page:    components.Page{Key: "seer_dashboard.RightCol"},
							Classes: "w-full xl:w-72 shrink-0",
							Children: []components.PageInterface{
								&components.FieldTitle{Getter: getters.Static("Workers")},
								&seerDashboardWorkerTabs{Page: components.Page{Key: "seer_dashboard.WorkerTabs"}},
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
