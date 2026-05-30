package p_seer_aisstream

import (
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/views"
)

var aisStreamMessageListQueryPatchers = views.QueryPatchers[AISStreamMessage]{
	{Key: "seer_aisstream.message_list.order", Value: views.QueryPatcherOrderBy[AISStreamMessage]{Order: "id DESC"}},
}

func init() {
	registerPluginView("seer_aisstream.MapView",
		lamu.GetPageView("seer_aisstream.MapPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}))

	registerPluginView("seer_aisstream.MessageListView",
		lamu.GetPageView("seer_aisstream.MessageTablePage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_aisstream.message.list", views.LayerList[AISStreamMessage]{
				Key:           getters.Static("aisStreamMessages"),
				PageSize:      getters.Static(uint(25)),
				QueryPatchers: aisStreamMessageListQueryPatchers,
			}))

	registerPluginView("seer_aisstream.MessageDetailView",
		lamu.GetPageView("seer_aisstream.MessageDetailPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_aisstream.message.detail", views.LayerDetail[AISStreamMessage]{
				Key:          getters.Static("aisStreamMessage"),
				PathParamKey: getters.Static("id"),
			}))
}
