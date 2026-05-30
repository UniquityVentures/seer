package p_seer_reddit

import (
	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
)

func init() {
	registerMenuPages()
	registerRedditSourcePages()
	registerRedditSourceCreatePages()
	registerRedditSourceUpdatePages()
	registerRedditPostPages()
	registerRedditRunnerPages()
	registerRedditRunnerWorkerPoolViews()
}

func registerMenuPages() {
	registerPluginPage("seer_reddit.RedditMenu", &components.SidebarMenu{
		Title: getters.Static("Reddit"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lamu.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Sources"),
				Url:   lamu.RoutePath("seer_reddit.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Posts"),
				Url:   lamu.RoutePath("seer_reddit.RedditPostListRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Workers"),
				Url:   lamu.RoutePath("seer_reddit.RedditRunnerListRoute", nil),
			},
		},
	})

	registerPluginPage("seer_reddit.RedditSourceDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Reddit source #%d", getters.Any(getters.Key[uint]("redditSource.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Sources"),
			Url:   lamu.RoutePath("seer_reddit.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lamu.RoutePath("seer_reddit.RedditSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditSource.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lamu.RoutePath("seer_reddit.RedditSourceUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditSource.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Posts"),
				Url: lamu.RoutePath("seer_reddit.RedditPostListBySourceRoute", map[string]getters.Getter[any]{
					"source_id": getters.Any(getters.Key[uint]("redditSource.ID")),
				}),
			},
		},
	})

	registerPluginPage("seer_reddit.RedditRunnerDetailMenu", &components.SidebarMenu{
		Title: getters.Key[string]("redditRunner.Name"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Workers"),
			Url:   lamu.RoutePath("seer_reddit.RedditRunnerListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lamu.RoutePath("seer_reddit.RedditRunnerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditRunner.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lamu.RoutePath("seer_reddit.RedditRunnerUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditRunner.ID")),
				}),
			},
		},
	})

	registerPluginPage("seer_reddit.RedditPostDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Post #%d", getters.Any(getters.Key[uint]("redditPost.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Posts"),
			Url:   lamu.RoutePath("seer_reddit.RedditPostListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lamu.RoutePath("seer_reddit.RedditPostDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditPost.ID")),
				}),
			},
		},
	})
}
