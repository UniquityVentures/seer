package p_seer_opensky

import (
	"context"
	"fmt"
	"time"

	"github.com/UniquityVentures/lamu/fields"
	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
)

type OpenSkyFlightTransition struct {
	ID             uint
	Icao24         string
	Callsign       string
	Latitude       float64
	Longitude      float64
	LastContact    int64
	Category       int
	PositionSource int
	FlightState    string // "takeoff" or "landed"
}

func (t *OpenSkyFlightTransition) Kind() string {
	return "opensky"
}

func (t *OpenSkyFlightTransition) IntelID() uint {
	return t.ID
}

func (t *OpenSkyFlightTransition) Content() string {
	return fmt.Sprintf("Flight %s (%s) transitioned to %s.\nPosition: (%f, %f)\nTime: %s\nCategory: %s\nPosition Source: %s",
		t.Callsign, t.Icao24, t.FlightState, t.Latitude, t.Longitude,
		time.Unix(t.LastContact, 0).UTC().Format(time.RFC3339),
		translateCategory(t.Category),
		translatePositionSource(t.PositionSource),
	)
}

func (t *OpenSkyFlightTransition) IntelDetail(ctx context.Context) (string, error) {
	return "", nil
}

func (t *OpenSkyFlightTransition) GetEvent(intel p_seer_intel.Intel) ([]p_seer_intel.IntelEvent, error) {
	return []p_seer_intel.IntelEvent{
		{
			IntelID:  intel.ID,
			Address:  fmt.Sprintf("Flight %s (%s) %s position", t.Callsign, t.Icao24, t.FlightState),
			Datetime: time.Unix(t.LastContact, 0).UTC(),
			Location: fields.NewPGPoint(t.Longitude, t.Latitude),
		},
	}, nil
}

func translateCategory(c int) string {
	switch c {
	case 0:
		return "No information at all"
	case 1:
		return "No ADS-B Emitter Category Information"
	case 2:
		return "Light (< 15500 lbs)"
	case 3:
		return "Small (15500 to 75000 lbs)"
	case 4:
		return "Large (75000 to 300000 lbs)"
	case 5:
		return "High Vortex Large (aircraft such as B-757)"
	case 6:
		return "Heavy (> 300000 lbs)"
	case 7:
		return "High Performance (> 5g acceleration and 400 kts)"
	case 8:
		return "Rotorcraft"
	case 9:
		return "Glider / sailplane"
	case 10:
		return "Lighter-than-air"
	case 11:
		return "Parachutist / Skydiver"
	case 12:
		return "Ultralight / hang-glider / paraglider"
	case 13:
		return "Reserved"
	case 14:
		return "Unmanned Aerial Vehicle"
	case 15:
		return "Space / Trans-atmospheric vehicle"
	case 16:
		return "Surface Vehicle – Emergency Vehicle"
	case 17:
		return "Surface Vehicle – Service Vehicle"
	case 18:
		return "Point Obstacle (includes tethered balloons)"
	case 19:
		return "Cluster Obstacle"
	case 20:
		return "Line Obstacle"
	default:
		return "Unknown"
	}
}

func translatePositionSource(ps int) string {
	switch ps {
	case 0:
		return "ADS-B"
	case 1:
		return "ASTERIX"
	case 2:
		return "MLAT"
	case 3:
		return "FLARM"
	default:
		return "Unknown"
	}
}
