package p_seer_gdelt

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/lago"
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
	return lago.RoutePath("seer_gdelt.EventDetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(strconv.FormatUint(uint64(e.ID), 10))),
	})(ctx)
}
