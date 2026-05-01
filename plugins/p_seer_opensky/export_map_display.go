package p_seer_opensky

import (
	"context"

	"gorm.io/gorm"
)

// MapDisplayPointWire is one CBOR marker for [github.com/UniquityVentures/lago/components.MapDisplay]
// (optional layer for merged dashboards).
type MapDisplayPointWire struct {
	Position  openSkyMapDisplayPosition `json:"position" cbor:"position"`
	Direction openSkyMapDisplayVector   `json:"direction,omitempty" cbor:"direction,omitempty"`
	Velocity  openSkyMapDisplayVector   `json:"velocity,omitempty" cbor:"velocity,omitempty"`
	Time      int64                     `json:"time,omitempty" cbor:"time,omitempty"`
	Link      string                    `json:"link,omitempty" cbor:"link,omitempty"`
	Layer     string                    `json:"layer,omitempty" cbor:"layer,omitempty"`
	Icon      string                    `json:"icon,omitempty" cbor:"icon,omitempty"`
	IconSize  float64                   `json:"iconSize,omitempty" cbor:"iconSize,omitempty"`
}

// MapDisplayPointsForBounds returns OpenSky-derived markers for the given viewport (degrees).
// If bounds is nil or invalid, loads the global aircraft set (same as the standalone map without viewport).
func MapDisplayPointsForBounds(ctx context.Context, db *gorm.DB, bounds *OpenSkyViewportBounds, layer string) ([]MapDisplayPointWire, error) {
	b := bounds
	if b != nil && !b.IsValid() {
		b = nil
	}
	aircraft, err := buildOpenSkyMapAircraft(ctx, db, b)
	if err != nil {
		return nil, err
	}
	inner := openSkyMapDisplayPoints(aircraft)
	out := make([]MapDisplayPointWire, len(inner))
	for i := range inner {
		out[i] = MapDisplayPointWire{
			Position:  inner[i].Position,
			Direction: inner[i].Direction,
			Velocity:  inner[i].Velocity,
			Time:      inner[i].Time,
			Link:      inner[i].Link,
			Layer:     layer,
		}
	}
	return out, nil
}
