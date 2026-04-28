package p_seer_assistant

import (
	"net/http"

	"github.com/UniquityVentures/lago/plugins/p_users"
	"github.com/UniquityVentures/lago/views"
	"gorm.io/gorm"
)

// assistantSessionUserScope restricts [SeerAssistantSession] rows to the signed-in user.
// Superusers see all sessions.
type assistantSessionUserScope struct{}

func (assistantSessionUserScope) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[SeerAssistantSession]) gorm.ChainInterface[SeerAssistantSession] {
	ctx := r.Context()
	u, ok := ctx.Value("$user").(p_users.User)
	if !ok {
		return query.Where("1 = 0")
	}
	if u.IsSuperuser {
		return query
	}
	return query.Where("user_id = ?", u.ID)
}
