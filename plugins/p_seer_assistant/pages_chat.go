package p_seer_assistant

import (
	"context"
	"fmt"
	"strings"

	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	. "maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
	. "maragu.dev/gomponents/html"
)

func registerAssistantMenuPages() {
	registerPluginPage("seer_assistant.AssistantMenu", &components.SidebarMenu{
		Title: getters.Static("Assistant"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lamu.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Chat"),
				Url:   lamu.RoutePath("seer_assistant.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("History"),
				Url:   lamu.RoutePath("seer_assistant.HistoryRoute", nil),
			},
		},
	})
}

type assistantChatRoot struct {
	components.Page
}

func (e *assistantChatRoot) Build(ctx context.Context) Node {
	sid := assistantOpenSessionID(ctx)
	wsPath := AppUrl + "ws/"
	if sid != 0 {
		wsPath = fmt.Sprintf("%s?session_id=%d", wsPath, sid)
	}
	hiddenVal := "0"
	if sid != 0 {
		hiddenVal = fmt.Sprintf("%d", sid)
	}
	transcriptInner := []Node{}
	if sid != 0 {
		nodes, err := assistantTranscriptNodes(ctx, sid)
		if err != nil {
			transcriptInner = append(transcriptInner, Div(Class("text-error text-sm"), Text("Could not load chat history")))
		} else if len(nodes) > 0 {
			transcriptInner = append(transcriptInner, Group(nodes))
		}
	}
	rootClass := "max-w-3xl mx-auto p-4 flex flex-col gap-4 min-h-[60vh]"
	transcriptClass := "flex flex-col gap-2 flex-1 overflow-y-auto border border-base-300 rounded-lg p-3 bg-base-200/40 min-h-[200px]"
	if e.Key == "assistant.SidebarChatInner" {
		rootClass = "max-w-3xl mx-auto p-0 flex flex-col gap-4 h-full overflow-hidden"
		transcriptClass = "flex flex-col gap-2 flex-1 overflow-y-auto border border-base-300 rounded-lg p-3 bg-base-200/40 min-h-0"
	}

	return Div(
		Class(rootClass),
		Attr("hx-ext", "ws"),
		Attr("ws-connect", wsPath),
		Script(Raw(`document.body.addEventListener("htmx:wsConfigSend", function(event) {
  if (!event || !event.detail || !event.detail.parameters) {
    return;
  }
  if (!event.target || event.target.id !== "seer_assistant_chat_form") {
    return;
  }
  var raw = event.detail.parameters.session_id;
  if (raw === undefined || raw === null || raw === "") {
    event.detail.parameters.session_id = 0;
    return;
  }
  var parsed = Number(raw);
  if (!Number.isNaN(parsed)) {
    event.detail.parameters.session_id = parsed;
  }
});
document.body.addEventListener("keydown", function(event) {
  if (!event.target || event.target.id !== "seer_assistant_chat_message") {
    return;
  }
  if (event.key !== "Enter" || event.shiftKey) {
    return;
  }
  event.preventDefault();
  var form = event.target.form;
  if (form) {
    form.requestSubmit();
  }
});
document.body.addEventListener("htmx:wsAfterSend", function(event) {
  if (!event.target || event.target.id !== "seer_assistant_chat_form") {
    return;
  }
  var ta = document.getElementById("seer_assistant_chat_message");
  var btn = document.getElementById("seer_assistant_chat_send");
  if (ta) {
    ta.value = "";
  }
  if (btn) {
    btn.disabled = true;
  }
});`)),
		Div(ID("seer_assistant_errors")),
		Div(
			ID("seer_assistant_transcript"),
			Class(transcriptClass),
			Group(transcriptInner),
		),
		Div(
			ID("seer_assistant_stream"),
			Class("min-h-[1.5rem] border border-dashed border-base-300 rounded p-2 text-sm"),
		),
		html.Form(ID("seer_assistant_chat_form"), Class("flex flex-col gap-2"), Attr("ws-send", ""), Input(ID("seer_assistant_session_id"), Type("hidden"), Name("session_id"), Value(hiddenVal)), Textarea(ID("seer_assistant_chat_message"), Name("message"), Class("textarea textarea-bordered w-full"), Rows("3"), Placeholder("Message…"), Required()), Button(ID("seer_assistant_chat_send"), Type("submit"), Class("btn btn-primary self-end"), Text("Send"))),
	)
}

func (e *assistantChatRoot) GetKey() string { return e.Key }

func (e *assistantChatRoot) GetRoles() []string { return e.Roles }

func registerAssistantChatPage() {
	registerPluginPage("seer_assistant.ChatPage", &components.ShellScaffold{
		Page: components.Page{Key: "seer_assistant.ChatPage"},
		Sidebar: []components.PageInterface{
			lamu.DynamicPage{Name: "seer_assistant.AssistantMenu"},
		},
		Children: []components.PageInterface{
			&assistantChatRoot{
				Page: components.Page{Key: "seer_assistant.ChatInner"},
			},
		},
	})
}

func assistantOpenSessionID(ctx context.Context) uint {
	if v := ctx.Value("assistantSession"); v != nil {
		if s, ok := v.(SeerAssistantSession); ok {
			return s.ID
		}
	}
	return 0
}

func assistantTranscriptNodes(ctx context.Context, sessionID uint) ([]Node, error) {
	if sessionID == 0 {
		return nil, nil
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}
	contents, err := LoadSessionContents(ctx, db, sessionID)
	if err != nil {
		return nil, err
	}
	out := make([]Node, 0, len(contents))
	for _, c := range contents {
		inner := strings.TrimSpace(assistantGenaiContentHTML(c))
		if inner == "" {
			continue
		}
		switch assistantTranscriptTurnKind(c) {
		case "assistant":
			out = append(out, assistantBubbleAssistantHTML(inner))
		case "tool":
			out = append(out, assistantBubbleToolHTML(inner))
		default:
			out = append(out, assistantBubbleUserHTML(inner))
		}
	}
	return out, nil
}

// assistantBubble*HTML: inner is from assistantGenaiContentHTML (escaped leaves); use Raw, not Text.
func assistantBubbleUserHTML(inner string) Node {
	return Div(
		Class("chat chat-end mb-2"),
		Div(Class("chat-header text-xs opacity-70"), Text("You")),
		Div(Class("chat-bubble text-sm chat-bubble-primary"), Raw(inner)),
	)
}

func assistantBubbleAssistantHTML(inner string) Node {
	return Div(
		Class("chat chat-start mb-2"),
		Div(Class("chat-header text-xs opacity-70"), Text("Assistant")),
		Div(Class("chat-bubble text-sm chat-bubble-secondary"), Raw(inner)),
	)
}

func assistantBubbleToolHTML(inner string) Node {
	return Div(
		Class("chat chat-start mb-2 w-full"),
		Div(Class("chat-header text-xs opacity-70"), Text("Tool")),
		El("details",
			Class("collapse collapse-arrow bg-base-200 border border-base-300 rounded-lg text-sm max-w-full"),
			El("summary", Class("collapse-title font-medium cursor-pointer pr-12"), Text("Tool Execution")),
			Div(Class("collapse-content p-3 pt-0 overflow-x-auto"), Raw(inner)),
		),
	)
}

type historySidebarPanel struct {
	components.Page
}

func (e *historySidebarPanel) Build(ctx context.Context) Node {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return Div(Class("text-error"), Text("Error: no database context"))
	}

	var sessions []SeerAssistantSession
	if err := db.Order("updated_at desc").Find(&sessions).Error; err != nil {
		return Div(Class("text-error"), Text("Error loading sessions"))
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

	currentSessionID := assistantOpenSessionID(ctx)
	var initialChatContent Node = Group{}
	var activeSessionName string

	if currentSessionID != 0 {
		for _, s := range sessions {
			if s.ID == currentSessionID {
				activeSessionName = strings.TrimSpace(s.Title)
				break
			}
		}
		if activeSessionName == "" {
			activeSessionName = fmt.Sprintf("Session #%d", currentSessionID)
		}

		chatInterface := components.Render(&assistantChatRoot{
			Page: components.Page{Key: "assistant.SidebarChatInner"},
		}, ctx)

		initialChatContent = Div(
			Class("flex-1 overflow-hidden min-h-0"),
			chatInterface,
		)
	} else {
		initialChatContent = Div(
			Class("flex-1 overflow-hidden min-h-0"),
			Attr("hx-push-url", "false"),
		)
	}

	xData := fmt.Sprintf(`{
		showModal: false,
		activeSessionId: $persist(0).as('assistant-sidebar-active-session-id'),
		init() {
			const serverSessionId = %d;
			if (serverSessionId !== 0) {
				this.activeSessionId = serverSessionId;
			} else {
				this.$nextTick(() => {
					if (this.activeSessionId !== 0) {
						const targetEl = document.getElementById('sidebar-chat-container');
						if (targetEl) {
							htmx.ajax('GET', '/seer-assistant/sidebar-chat/' + this.activeSessionId + '/', {
								target: targetEl,
								swap: 'innerHTML',
								source: targetEl
							});
						}
					}
				});
			}
		}
	}`, currentSessionID)

	return Div(
		Attr("x-data", xData),
		Attr("@new-session-created.window", "activeSessionId = $event.detail.id; showModal = false; htmx.ajax('GET', '/seer-assistant/sidebar-chat/' + activeSessionId + '/', {target: '#sidebar-chat-container', swap: 'innerHTML', source: $el})"),
		Class("flex flex-col gap-0 p-2 h-full overflow-hidden"),
		Attr("hx-push-url", "false"),

		// Header Row: Session name on left, buttons (History, New Chat) on right
		Div(
			Class("flex justify-between items-center flex-none border-b border-base-300 pb-2 px-1"),
			Div(
				ID("session-name-container"),
				Class("text-sm font-semibold truncate max-w-[70%]"),
				Text(activeSessionName),
			),
			Div(
				Class("flex gap-1 flex-none"),
				// History Button
				Button(
					Class("btn btn-sm btn-ghost btn-circle"),
					Attr("@click", "showModal = true"),
					components.Render(components.Icon{Name: "clock"}, ctx),
				),
				// New Chat Button
				Button(
					Class("btn btn-sm btn-ghost btn-circle"),
					Attr("hx-post", "/seer-assistant/new-session/"),
					Attr("hx-swap", "none"),
					Attr("hx-push-url", "false"),
					components.Render(components.Icon{Name: "plus"}, ctx),
				),
			),
		),

		// Selected Session Name & Chat under the button (swapped dynamically)
		Div(
			ID("sidebar-chat-container"),
			Class("flex-1 flex flex-col gap-4 overflow-hidden min-h-0"),
			Attr("hx-push-url", "false"),
			initialChatContent,
		),

		// Custom Modal using standard dialog element, controlled by Alpine
		El("dialog",
			Attr("x-show", "showModal"),
			Attr(":class", "showModal ? 'modal modal-open' : 'modal'"),
			Div(
				Class("modal-box bg-base-100 max-w-lg border border-base-300 p-6 relative"),
				// Close button
				Button(
					Type("button"),
					Class("btn btn-sm btn-circle btn-ghost absolute right-3 top-3"),
					Attr("@click", "showModal = false"),
					components.Render(components.Icon{Name: "x-mark"}, ctx),
				),
				// Modal Title
				H3(Class("text-lg font-bold mb-4"), Text("Conversations")),
				// Sessions List
				Div(
					ID("modal-sessions-list"),
					Class("max-h-60 overflow-y-auto flex flex-col bg-base-200 rounded border border-base-300"),
					Group(sessionItems),
				),
			),
			// Backdrop clicking closes the modal
			FormEl(
				Method("dialog"),
				Class("modal-backdrop"),
				Button(Attr("@click", "showModal = false"), Text("close")),
			),
		),
	)
}

func (e *historySidebarPanel) GetKey() string     { return e.Key }
func (e *historySidebarPanel) GetRoles() []string { return e.Roles }

// sidebarChatPage is rendered dynamically inside the sidebar container when a session is switched.
type sidebarChatPage struct {
	components.Page
}

func (e *sidebarChatPage) Build(ctx context.Context) Node {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return Div(Class("text-error"), Text("Error: no database context"))
	}
	currentSessionID := assistantOpenSessionID(ctx)
	if currentSessionID == 0 {
		return Div(Class("text-error"), Text("No session selected"))
	}
	var session SeerAssistantSession
	if err := db.First(&session, currentSessionID).Error; err != nil {
		return Div(Class("text-error"), Text("Session not found"))
	}

	title := strings.TrimSpace(session.Title)
	if title == "" {
		title = fmt.Sprintf("Session #%d", session.ID)
	}

	chatInterface := components.Render(&assistantChatRoot{
		Page: components.Page{Key: "assistant.SidebarChatInner"},
	}, ctx)

	return Group{
		Div(
			ID("session-name-container"),
			Attr("hx-swap-oob", "true"),
			Class("text-sm font-semibold truncate max-w-[70%]"),
			Text(title),
		),
		Div(
			Class("flex-1 overflow-hidden min-h-0"),
			chatInterface,
		),
	}
}

func (e *sidebarChatPage) GetKey() string     { return e.Key }
func (e *sidebarChatPage) GetRoles() []string { return e.Roles }

func sidebarChatPageLookup(name string) (components.PageInterface, bool) {
	if name == "seer_assistant.SidebarChatPage" {
		return &sidebarChatPage{
			Page: components.Page{Key: "seer_assistant.SidebarChatPage"},
		}, true
	}
	return nil, false
}

func init() {
	registerAssistantMenuPages()
	registerAssistantChatPage()

	components.RegistryRightSidebar.Register("assistant.history_panel", components.SidebarItem{
		Icon: "clock",
		Content: &historySidebarPanel{
			Page: components.Page{Key: "assistant.history_panel"},
		},
	})
	// Trigger air rebuild to compile with newly added replace directives
}
