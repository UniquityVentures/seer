package p_seer_dashboard

import (
	"context"
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"

	"github.com/UniquityVentures/lago/components"
	"github.com/UniquityVentures/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// seerDashboardMapDefaultSelectZoom is the zoom level applied when a region is picked.
const seerDashboardMapDefaultSelectZoom = 8.0

// seerDashboardMapIDSanitize mirrors lago/components.mapDisplayIDSanitize so the id suffix
// derived here matches the one MapDisplay computes from the same Page.Key.
var seerDashboardMapIDSanitize = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

// SeerDashboardMap wraps [components.MapDisplay] with a "select a region first" overlay.
//
// When rendered, the user sees a world map covered by a dimmed overlay that reads
// "Click anywhere to select a region". On the first click the wrapper:
//
//  1. converts the click point to lng/lat via the underlying MapLibre instance,
//  2. calls flyTo on that center at SelectZoom (default 8),
//  3. opens the data WebSocket so MapDisplay starts streaming markers, and
//  4. hides the overlay.
//
// The wrapped MapDisplay is started in deferred mode (DeferStart=true), so no
// WebSocket traffic happens until a region has been picked. After selection the
// behavior is identical to a plain MapDisplay; subsequent pan/zoom continue to
// drive viewport-bounded server queries.
type SeerDashboardMap struct {
	components.Page
	// DataURL is forwarded to the wrapped MapDisplay (ws/wss URL or same-origin path).
	DataURL getters.Getter[string]
	// RefreshMS is forwarded to the wrapped MapDisplay (reconnect delay in ms).
	RefreshMS getters.Getter[int64]
	// SelectZoom is the MapLibre zoom level applied after the user picks a region.
	// When nil, defaults to 8.
	SelectZoom getters.Getter[float64]
	// Classes for the map container div (passed through to MapDisplay).
	Classes string
}

func (e *SeerDashboardMap) GetKey() string     { return e.Key }
func (e *SeerDashboardMap) GetRoles() []string { return e.Roles }

// seerDashboardMapIDSuffix sanitizes a Page.Key the same way MapDisplay does so we can
// address window["mapDisplay_<suffix>"] and the deterministic map element id.
func seerDashboardMapIDSuffix(pageKey string) string {
	s := strings.TrimSpace(pageKey)
	if s == "" {
		return "default"
	}
	s = seerDashboardMapIDSanitize.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "default"
	}
	if len(s) > 48 {
		s = s[:48]
	}
	return s
}

func (e *SeerDashboardMap) Build(ctx context.Context) Node {
	selectZoom := seerDashboardMapDefaultSelectZoom
	if e.SelectZoom != nil {
		v, err := e.SelectZoom(ctx)
		if err != nil {
			slog.Error("SeerDashboardMap SelectZoom getter failed", "error", err, "key", e.Key)
			return components.ContainerError{
				Page:  components.Page{Key: e.Key + ".err"},
				Error: getters.Static(err),
			}.Build(ctx)
		}
		selectZoom = v
	}

	innerKey := e.Key + ".map"
	innerSuffix := seerDashboardMapIDSuffix(innerKey)
	mapElID := "mapdisplay-" + innerSuffix + "-map"
	overlayID := "seerdash-" + innerSuffix + "-overlay"

	classes := strings.TrimSpace(e.Classes)
	wrapperClasses := "relative w-full"

	innerSuffixBytes, _ := json.Marshal(innerSuffix)
	mapElIDBytes, _ := json.Marshal(mapElID)
	overlayIDBytes, _ := json.Marshal(overlayID)
	selectZoomBytes, _ := json.Marshal(selectZoom)

	mapDisplay := &components.MapDisplay{
		Page:              components.Page{Key: innerKey},
		DataURL:           e.DataURL,
		RefreshMS:         e.RefreshMS,
		Classes:           classes,
		DeferStart:        getters.Static(true),
		SkipAutoFitBounds: getters.Static(true),
	}

	initJS := `(function(){
  var suffix = ` + string(innerSuffixBytes) + `;
  var mapElId = ` + string(mapElIDBytes) + `;
  var overlayId = ` + string(overlayIDBytes) + `;
  var selectZoom = ` + string(selectZoomBytes) + `;

  var armed = false;
  var pollAttempts = 0;

  function api() { return window["mapDisplay_" + suffix]; }

  function arm() {
    if (armed) { return; }
    var overlay = document.getElementById(overlayId);
    var mapEl = document.getElementById(mapElId);
    var a = api();
    if (!overlay || !mapEl || !a || typeof a.unproject !== "function") { return; }
    armed = true;
    overlay.addEventListener("click", function (ev) {
      ev.preventDefault();
      ev.stopPropagation();
      var rect = mapEl.getBoundingClientRect();
      var x = ev.clientX - rect.left;
      var y = ev.clientY - rect.top;
      var ll = a.unproject(x, y);
      if (!ll || typeof ll.lng !== "number" || typeof ll.lat !== "number") { return; }
      a.flyTo(ll.lng, ll.lat, selectZoom);
      a.start();
      overlay.style.display = "none";
    }, { once: true });
  }

  function onReady(ev) {
    if (!ev || !ev.detail || ev.detail.suffix !== suffix) { return; }
    arm();
  }

  document.addEventListener("mapDisplayReady", onReady);

  function pollArm() {
    if (armed) { return; }
    var a = api();
    if (a && typeof a.isReady === "function" && a.isReady()) {
      arm();
      return;
    }
    pollAttempts++;
    if (pollAttempts > 200) { return; }
    setTimeout(pollArm, 50);
  }
  pollArm();
})();`

	return Group([]Node{
		Div(
			ID("seerdash-"+innerSuffix+"-wrap"),
			Class(wrapperClasses),
			components.Render(mapDisplay, ctx),
			Div(
				ID(overlayID),
				Class("absolute inset-0 z-[2] flex items-center justify-center cursor-crosshair rounded-box"),
				Span(
					Class("badge badge-lg badge-neutral shadow-lg pointer-events-none"),
					Text("Click anywhere to select a region"),
				),
			),
		),
		Script(Raw(initJS)),
	})
}
