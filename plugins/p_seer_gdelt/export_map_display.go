package p_seer_gdelt

import (
	"context"

	"github.com/UniquityVentures/lago/getters"
	"gorm.io/gorm"
)

// MapDisplayPointWire is one CBOR marker for [github.com/UniquityVentures/lago/components.MapDisplay]
// (GDELT events are undirected points; optional layer for merged dashboards).
type MapDisplayPointWire struct {
	Position struct {
		Lat float64 `json:"lat" cbor:"lat"`
		Lng float64 `json:"lng" cbor:"lng"`
	} `json:"position" cbor:"position"`
	Link  string `json:"link,omitempty" cbor:"link,omitempty"`
	Title string `json:"title,omitempty" cbor:"title,omitempty"`
	Layer string `json:"layer,omitempty" cbor:"layer,omitempty"`
}

func pointInMapViewport(lat, lng, west, south, east, north float64) bool {
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

// MapDisplayPointsForBounds returns GDELT-derived markers (same cap as the GDELT map layer),
// optionally filtered to a lat/lng viewport when south <= north.
func MapDisplayPointsForBounds(ctx context.Context, _ *gorm.DB, west, south, east, north float64, layer string) ([]MapDisplayPointWire, error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}
	events, err := gorm.G[Event](db.WithContext(ctx)).
		Order("id DESC").
		Limit(gdeltMapMaxEvents).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	markers := buildGDELTMapMarkers(events)
	out := make([]MapDisplayPointWire, 0, len(markers))
	for _, m := range markers {
		if !pointInMapViewport(m.Lat, m.Lng, west, south, east, north) {
			continue
		}
		var p MapDisplayPointWire
		p.Position.Lat = m.Lat
		p.Position.Lng = m.Lng
		p.Link = m.DetailPath
		p.Title = m.Title
		p.Layer = layer
		out = append(out, p)
	}
	return out, nil
}
