package p_seer_gdelt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/UniquityVentures/lamu/fields"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
)

const gdeltIntelKind = "gdelt"

// Kind satisfies [p_seer_intel.IntelKind] for [Event].
func (e Event) Kind() string {
	return gdeltIntelKind
}

// IntelID satisfies [p_seer_intel.IntelKind] for [Event].
func (e Event) IntelID() uint {
	return e.ID
}

// Content satisfies [p_seer_intel.IntelKind] for [*Event]: markdown from persisted GDELT fields.
func (e *Event) Content() string {
	if e == nil {
		return ""
	}
	var b strings.Builder
	title := strings.TrimSpace(e.Actor1Name)
	if t2 := strings.TrimSpace(e.Actor2Name); t2 != "" {
		if title != "" {
			title = title + " / " + t2
		} else {
			title = t2
		}
	}
	if title != "" {
		b.WriteString("# ")
		b.WriteString(title)
		b.WriteString("\n\n")
	}
	if e.SQLDate != 0 {
		s := strconv.Itoa(e.SQLDate)
		if len(s) == 8 {
			fmt.Fprintf(&b, "**Date:** %s-%s-%s\n\n", s[:4], s[4:6], s[6:])
		}
	}
	if c := strings.TrimSpace(e.EventCode); c != "" {
		fmt.Fprintf(&b, "**Event code:** %s\n", c)
	}
	if c := strings.TrimSpace(e.ActionGeoFullName); c != "" {
		fmt.Fprintf(&b, "**Action location:** %s (%s)\n", c, strings.TrimSpace(e.ActionGeoCountryCode))
	}
	if e.NumMentions != 0 {
		fmt.Fprintf(&b, "**Mentions:** %d\n", e.NumMentions)
	}
	if e.AvgTone != 0 {
		fmt.Fprintf(&b, "**Avg tone:** %.3f\n", e.AvgTone)
	}
	if u := strings.TrimSpace(e.SourceURL); u != "" {
		fmt.Fprintf(&b, "\n**Source:** %s\n", u)
	}
	if e.GlobalEventID != 0 {
		fmt.Fprintf(&b, "\n**GLOBALEVENTID:** %d\n", e.GlobalEventID)
	}
	return strings.TrimSpace(b.String())
}

// IntelDetail satisfies [p_seer_intel.IntelKind] for [*Event].
func (e *Event) IntelDetail(ctx context.Context) (string, error) {
	if e == nil || e.ID == 0 {
		return "", fmt.Errorf("p_seer_gdelt: IntelDetail: missing event")
	}
	return lamu.RoutePath("seer_gdelt.EventDetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(strconv.FormatUint(uint64(e.ID), 10))),
	})(ctx)
}

// GetEvent satisfies [p_seer_intel.IntelEventKind] for [*Event].
func (e *Event) GetEvent(intel p_seer_intel.Intel) ([]p_seer_intel.IntelEvent, error) {
	var dt time.Time
	if e.SQLDate != 0 {
		s := strconv.Itoa(e.SQLDate)
		if len(s) == 8 {
			year, _ := strconv.Atoi(s[:4])
			month, _ := strconv.Atoi(s[4:6])
			day, _ := strconv.Atoi(s[6:])
			dt = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		}
	}
	if dt.IsZero() {
		dt = time.Now().UTC()
	}

	var events []p_seer_intel.IntelEvent

	// 1. Actor 1 Geo
	if strings.TrimSpace(e.Actor1GeoFullName) != "" && (e.Actor1GeoLat != 0 || e.Actor1GeoLong != 0) {
		events = append(events, p_seer_intel.IntelEvent{
			IntelID:  intel.ID,
			Address:  e.Actor1GeoFullName,
			Datetime: dt,
			Location: fields.NewPGPoint(e.Actor1GeoLong, e.Actor1GeoLat),
		})
	}

	// 2. Actor 2 Geo
	if strings.TrimSpace(e.Actor2GeoFullName) != "" && (e.Actor2GeoLat != 0 || e.Actor2GeoLong != 0) {
		events = append(events, p_seer_intel.IntelEvent{
			IntelID:  intel.ID,
			Address:  e.Actor2GeoFullName,
			Datetime: dt,
			Location: fields.NewPGPoint(e.Actor2GeoLong, e.Actor2GeoLat),
		})
	}

	// 3. Action Geo
	if strings.TrimSpace(e.ActionGeoFullName) != "" && e.ActionGeoPoint.Valid {
		events = append(events, p_seer_intel.IntelEvent{
			IntelID:  intel.ID,
			Address:  e.ActionGeoFullName,
			Datetime: dt,
			Location: e.ActionGeoPoint,
		})
	}

	return events, nil
}
