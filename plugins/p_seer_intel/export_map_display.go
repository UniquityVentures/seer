package p_seer_intel

import (
	"context"
	"strconv"
	"strings"

	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/lago"
	"gorm.io/gorm"
)

// IntelEventsTable is the GORM/Postgres table for [IntelEvent].
const IntelEventsTable = "intel_events"

// IntelsTable is the GORM/Postgres table for [Intel].
const IntelsTable = "intels"

// intelMapMaxMarkers caps dashboard map payloads for intel-derived points.
const intelMapMaxMarkers = 5000

// MapDisplayPointWire is one CBOR marker for [github.com/UniquityVentures/lago/components.MapDisplay]
// (intel events with geocoded locations; optional layer for merged dashboards).
type MapDisplayPointWire struct {
	Position struct {
		Lat float64 `json:"lat" cbor:"lat"`
		Lng float64 `json:"lng" cbor:"lng"`
	} `json:"position" cbor:"position"`
	Link  string `json:"link,omitempty" cbor:"link,omitempty"`
	Title string `json:"title,omitempty" cbor:"title,omitempty"`
	Layer string `json:"layer,omitempty" cbor:"layer,omitempty"`
	Color string `json:"color,omitempty" cbor:"color,omitempty"`
}

func intelPointInMapViewport(lat, lng, west, south, east, north float64) bool {
	if south > north {
		return true
	}
	if lat < south || lat > north {
		return false
	}
	if west <= east {
		return lng >= west && lng <= east
	}
	return lng >= west || lng <= east
}

// MapDisplayPointsForBounds returns markers for [IntelEvent] rows that have a geocoded [IntelEvent.Location],
// joined to a non–soft-deleted [Intel]. When south > north (sentinel), no lat/lng filtering is applied.
func MapDisplayPointsForBounds(ctx context.Context, _ *gorm.DB, west, south, east, north float64, layer string) ([]MapDisplayPointWire, error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var rows []IntelEvent
	err = db.WithContext(ctx).Model(&IntelEvent{}).
		Preload("Intel").
		Joins("INNER JOIN " + IntelsTable + " ON " + IntelsTable + ".id = " + IntelEventsTable + ".intel_id AND " + IntelsTable + ".deleted_at IS NULL").
		Where(IntelEventsTable + ".deleted_at IS NULL").
		Order(IntelEventsTable + ".id DESC").
		Limit(intelMapMaxMarkers).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]MapDisplayPointWire, 0, len(rows))
	for _, ie := range rows {
		if !ie.Location.Valid {
			continue
		}
		lat := ie.Location.P.Y
		lng := ie.Location.P.X
		if !intelPointInMapViewport(lat, lng, west, south, east, north) {
			continue
		}
		var p MapDisplayPointWire
		p.Position.Lat = lat
		p.Position.Lng = lng
		p.Link = intelDetailPath(ctx, ie.IntelID)
		if ie.Intel != nil {
			p.Title = strings.TrimSpace(ie.Intel.Title)
		}
		p.Layer = layer
		p.Color = "#7c3aed"
		out = append(out, p)
	}
	return out, nil
}

func intelDetailPath(ctx context.Context, intelID uint) string {
	if intelID == 0 {
		return ""
	}
	href, err := lago.RoutePath("seer_intel.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(strconv.FormatUint(uint64(intelID), 10))),
	})(ctx)
	if err != nil || href == "" {
		return ""
	}
	return href
}
