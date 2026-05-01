package p_seer_aisstream

import (
	"context"

	"gorm.io/gorm"
)

// MapDisplayPointWire is one CBOR marker for [github.com/UniquityVentures/lago/components.MapDisplay]
// (optional layer for merged dashboards).
type MapDisplayPointWire struct {
	Position  aisStreamMapDisplayPosition `json:"position" cbor:"position"`
	Direction aisStreamMapDisplayVector   `json:"direction,omitempty" cbor:"direction,omitempty"`
	Time      int64                       `json:"time,omitempty" cbor:"time,omitempty"`
	Link      string                      `json:"link,omitempty" cbor:"link,omitempty"`
	Layer     string                      `json:"layer,omitempty" cbor:"layer,omitempty"`
	Icon      string                      `json:"icon,omitempty" cbor:"icon,omitempty"`
	IconSize  float64                     `json:"iconSize,omitempty" cbor:"iconSize,omitempty"`
}

// MapDisplayPointsForBounds returns AISStream-derived markers for the given viewport (degrees).
// If bounds is nil or invalid, the query is unbounded (subject to plugin SQL limits).
func MapDisplayPointsForBounds(ctx context.Context, db *gorm.DB, bounds *AisstreamViewportBounds, layer string) ([]MapDisplayPointWire, error) {
	b := bounds
	if b != nil && !b.IsValid() {
		b = nil
	}
	vessels, err := buildAISStreamMapVessels(ctx, db, b)
	if err != nil {
		return nil, err
	}
	inner := aisStreamMapDisplayPoints(vessels)
	out := make([]MapDisplayPointWire, len(inner))
	for i := range inner {
		out[i] = MapDisplayPointWire{
			Position:  inner[i].Position,
			Direction: inner[i].Direction,
			Time:      inner[i].Time,
			Link:      inner[i].Link,
			Layer:     layer,
		}
	}
	return out, nil
}
