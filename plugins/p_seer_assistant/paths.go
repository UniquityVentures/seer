package p_seer_assistant

import (
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"golang.org/x/net/websocket"
)

func init() {
	registerPluginRoute("seer_assistant.DefaultRoute", lamu.Route{
		Path:    AppUrl,
		Handler: lamu.NewDynamicView("seer_assistant.ChatView"),
	})

	registerPluginRoute("seer_assistant.HistoryRoute", lamu.Route{
		Path:    AppUrl + "history/",
		Handler: lamu.NewDynamicView("seer_assistant.HistoryView"),
	})

	registerPluginRoute("seer_assistant.ChatSessionRoute", lamu.Route{
		Path:    AppUrl + "c/{id}/",
		Handler: lamu.NewDynamicView("seer_assistant.ChatSessionView"),
	})

	registerPluginRoute("seer_assistant.WSRoute", lamu.Route{
		Path: AppUrl + "ws/",
		Handler: p_users.RequireAuth(websocket.Server{
			Handler: assistantWebSocketConn,
		}),
	})
}
