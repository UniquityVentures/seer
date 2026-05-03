package p_seer_gdelt

import (
	"context"
	"fmt"
	"log"

	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
	"gorm.io/gorm"
)

func loadGDELTEventIntelKind(ctx context.Context, db *gorm.DB, id uint) (p_seer_intel.IntelKind, error) {
	if db == nil {
		return nil, fmt.Errorf("p_seer_gdelt: loadGDELTEventIntelKind: nil db")
	}
	var ev Event
	if err := db.WithContext(ctx).First(&ev, id).Error; err != nil {
		return nil, err
	}
	return &ev, nil
}

func init() {
	if err := p_seer_intel.RegistryIntelKind.Register((Event{}).Kind(), loadGDELTEventIntelKind); err != nil {
		log.Panic(err)
	}
}
