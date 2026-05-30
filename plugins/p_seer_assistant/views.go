package p_seer_assistant

import (
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/views"
)

func init() {
	sessionListPatchers := views.QueryPatchers[SeerAssistantSession]{
		{Key: "seer_assistant.session.user_scope", Value: assistantSessionUserScope{}},
		{Key: "seer_assistant.session.order", Value: views.QueryPatcherOrderBy[SeerAssistantSession]{Order: "updated_at DESC"}},
	}
	sessionDetailPatchers := views.QueryPatchers[SeerAssistantSession]{
		{Key: "seer_assistant.session.user_scope", Value: assistantSessionUserScope{}},
	}

	registerPluginView("seer_assistant.ChatView",
		lamu.GetPageView("seer_assistant.ChatPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}))

	registerPluginView("seer_assistant.HistoryView",
		lamu.GetPageView("seer_assistant.HistoryPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_assistant.session.list", views.LayerList[SeerAssistantSession]{
				Key:           getters.Static("assistantSessions"),
				QueryPatchers: sessionListPatchers,
			}))

	registerPluginView("seer_assistant.ChatSessionView",
		lamu.GetPageView("seer_assistant.ChatPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_assistant.session.detail", views.LayerDetail[SeerAssistantSession]{
				Key:           getters.Static("assistantSession"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: sessionDetailPatchers,
			}))
}
