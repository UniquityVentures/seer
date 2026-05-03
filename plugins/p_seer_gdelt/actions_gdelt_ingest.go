package p_seer_gdelt

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

const gdeltIntelParallelism = 8

func createIntelForGDELTEventIfMissing(ctx context.Context, db *gorm.DB, ev Event) error {
	if ev.ID == 0 {
		return nil
	}
	kind := (Event{}).Kind()
	exists, err := p_seer_intel.IntelExistsForSource(ctx, db, kind, ev.ID)
	if err != nil {
		return fmt.Errorf("exists check: %w", err)
	}
	if exists {
		return nil
	}
	intel, err := p_seer_intel.NewFromIntelKind(ctx, &ev)
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	if err := p_seer_intel.CreateIntelAndEvent(ctx, db, &intel); err != nil {
		return fmt.Errorf("persist: %w", err)
	}
	return nil
}

// RunGDELTEventsIntelIngest runs [createIntelForGDELTEventIfMissing] for each event with bounded parallelism.
func RunGDELTEventsIntelIngest(ctx context.Context, db *gorm.DB, events []Event) {
	if db == nil || len(events) == 0 {
		return
	}
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(gdeltIntelParallelism)
	for i := range events {
		ev := events[i]
		if ev.ID == 0 {
			continue
		}
		g.Go(func() error {
			if err := createIntelForGDELTEventIfMissing(ctx, db, ev); err != nil {
				slog.Warn("p_seer_gdelt: add intel event", "event_id", ev.ID, "error", err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		slog.Warn("p_seer_gdelt: intel ingest group", "error", err)
	}
}
