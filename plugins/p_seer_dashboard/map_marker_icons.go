package p_seer_dashboard

import (
	"net/http"

	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/lago/plugins/p_users"
)

// Tiny same-origin SVG markers for MapDisplay (directed symbols rotate with bearing).
// Both use the same upward arrow; white stroke layer reads as a border on map tiles.
var (
	mapMarkerOpenSkySVG   = []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="#fff" stroke-width="2.25" stroke-linejoin="round" d="M12 3l9 14h-6l-3-5-3 5H3L12 3z"/><path fill="#0369a1" d="M12 3l9 14h-6l-3-5-3 5H3L12 3z"/></svg>`)
	mapMarkerAISStreamSVG = []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="#fff" stroke-width="2.25" stroke-linejoin="round" d="M12 3l9 14h-6l-3-5-3 5H3L12 3z"/><path fill="#0f766e" d="M12 3l9 14h-6l-3-5-3 5H3L12 3z"/></svg>`)
)

func registerDashboardMarkerIconRoutes() {
	_ = lago.RegistryRoute.Register("seer_dashboard.MapMarkerOpenSkyRoute", lago.Route{
		Path:    AppUrl + "map/icon/opensky.svg",
		Handler: p_users.RequireAuth(serveMapMarkerSVG(mapMarkerOpenSkySVG)),
	})
	_ = lago.RegistryRoute.Register("seer_dashboard.MapMarkerAISStreamRoute", lago.Route{
		Path:    AppUrl + "map/icon/aisstream.svg",
		Handler: p_users.RequireAuth(serveMapMarkerSVG(mapMarkerAISStreamSVG)),
	})
}

func serveMapMarkerSVG(body []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(body)
	})
}
