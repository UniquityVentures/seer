package p_seer_reddit

import (
	"context"
	"net/http"
	"strconv"

	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// redditPostToolbarBusyGetter disables Reddit post toolbar POSTs while async source fetch runs.
func redditPostToolbarBusyGetter() getters.Getter[Node] {
	return func(context.Context) (Node, error) {
		if redditFetchPostsActive.Load() {
			return Group{Disabled(), Class("btn-disabled")}, nil
		}
		return nil, nil
	}
}

func redditPostListViewPollURL(ctx context.Context, bySource bool, sourceID uint) (string, error) {
	routeName := "seer_reddit.RedditPostListRoute"
	var pathArgs map[string]getters.Getter[any]
	if bySource {
		routeName = "seer_reddit.RedditPostListBySourceRoute"
		pathArgs = map[string]getters.Getter[any]{
			"source_id": getters.Any(getters.Static(strconv.FormatUint(uint64(sourceID), 10))),
		}
	}
	base, err := lamu.RoutePath(routeName, pathArgs)(ctx)
	if err != nil {
		return "", err
	}
	reqVal := ctx.Value("$request")
	r, ok := reqVal.(*http.Request)
	if !ok || r == nil || r.URL == nil || r.URL.RawQuery == "" {
		return base, nil
	}
	return base + "?" + r.URL.RawQuery, nil
}

func redditPostListTableShellGetter(bySource bool) getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		var sourceID uint
		if bySource {
			sid, err := getters.Key[uint]("redditSource.ID")(ctx)
			if err != nil {
				return nil, err
			}
			sourceID = sid
		}
		tbl := newRedditPostDataTable()
		busy := false
		if bySource {
			busy = redditFetchPostsActive.Load()
		}
		if !busy {
			return tbl, nil
		}
		u, err := redditPostListViewPollURL(ctx, bySource, sourceID)
		if err != nil {
			return nil, err
		}
		return &components.HTMXPolling{
			Page:     components.Page{Key: "seer_reddit.RedditPostTablePolling"},
			URL:      getters.Static(u),
			Children: []components.PageInterface{tbl},
		}, nil
	}
}
