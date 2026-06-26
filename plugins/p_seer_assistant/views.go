package p_seer_assistant

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/registry"
	"github.com/UniquityVentures/lamu/views"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
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

	registerPluginView("seer_assistant.SidebarChatView",
		&views.View{
			PageName:   "seer_assistant.SidebarChatPage",
			PageLookup: sidebarChatPageLookup,
			Layers: []registry.Pair[string, views.Layer]{
				{Key: "p_users.auth", Value: p_users.AuthenticationLayer{}},
				{Key: "seer_assistant.session.detail", Value: views.LayerDetail[SeerAssistantSession]{
					Key:           getters.Static("assistantSession"),
					PathParamKey:  getters.Static("id"),
					QueryPatchers: sessionDetailPatchers,
				}},
			},
		})
}

func handleNewSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("$user").(p_users.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		http.Error(w, "No database connection", http.StatusInternalServerError)
		return
	}
	session, err := CreateSession(ctx, db, user.ID)
	if err != nil {
		http.Error(w, "Could not create session", http.StatusInternalServerError)
		return
	}

	// Load all sessions to build the updated sessions list for OOB swap
	var sessions []SeerAssistantSession
	if err := db.Order("updated_at desc").Find(&sessions).Error; err != nil {
		http.Error(w, "Could not load sessions", http.StatusInternalServerError)
		return
	}

	var sessionItems []Node
	for _, s := range sessions {
		title := strings.TrimSpace(s.Title)
		if title == "" {
			title = fmt.Sprintf("Session #%d", s.ID)
		}
		sessionItems = append(sessionItems, Div(
			Class("p-3 hover:bg-base-300 rounded cursor-pointer transition border-b border-base-300 last:border-b-0 text-sm block no-underline text-base-content"),
			Attr("hx-get", fmt.Sprintf("/seer-assistant/sidebar-chat/%d/", s.ID)),
			Attr("hx-target", "#sidebar-chat-container"),
			Attr("hx-swap", "innerHTML"),
			Attr("hx-push-url", "false"),
			Attr("@click", fmt.Sprintf("activeSessionId = %d; showModal = false", s.ID)),
			Text(title),
		))
	}

	if len(sessionItems) == 0 {
		sessionItems = []Node{
			Div(Class("p-4 text-center text-sm opacity-50"), Text("No sessions found")),
		}
	}

	// Render the updated sessions list as an OOB swap
	oobList := Div(
		ID("modal-sessions-list"),
		Attr("hx-swap-oob", "innerHTML"),
		Group(sessionItems),
	)

	// Set HX-Trigger header so Alpine.js updates activeSessionId and calls htmx.ajax to load the chat UI
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"new-session-created": {"id": %d}}`, session.ID))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_ = oobList.Render(w)
}
