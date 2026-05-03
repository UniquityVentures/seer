package p_seer_gdelt

import (
	"strings"
	"time"

	"github.com/UniquityVentures/lago/lago"
	"gorm.io/gorm"
)

const (
	EventsTable       = "seer_gdelt_events"
	GDELTWorkersTable = "seer_gdelt_workers"
	GDELTSourcesTable = "seer_gdelt_sources"
)

// GDELTWorker is a cadence bucket for scheduled BigQuery pulls; [GDELTSource.GDELTWorkerID] is optional.
type GDELTWorker struct {
	gorm.Model

	Name           string        `gorm:"size:64;not null;uniqueIndex"`
	Duration       time.Duration `gorm:"not null"`
	GDELTSourceIDs []uint        `gorm:"-"`
}

func (GDELTWorker) TableName() string {
	return GDELTWorkersTable
}

func gdeltWorkerFromUpdateDest(tx *gorm.DB) *GDELTWorker {
	if tx == nil || tx.Statement == nil {
		return nil
	}
	switch d := tx.Statement.Dest.(type) {
	case GDELTWorker:
		return &d
	case *GDELTWorker:
		if d == nil {
			return nil
		}
		return d
	default:
		return nil
	}
}

func (w *GDELTWorker) AfterSave(tx *gorm.DB) error {
	worker := gdeltWorkerFromUpdateDest(tx)
	if worker == nil {
		worker = w
	}
	if worker == nil || worker.ID == 0 {
		return nil
	}
	if err := tx.Model(&GDELTSource{}).Where("gdelt_worker_id = ?", worker.ID).Update("gdelt_worker_id", nil).Error; err != nil {
		return err
	}
	ids := worker.GDELTSourceIDs
	if len(ids) == 0 {
		return nil
	}
	return tx.Model(&GDELTSource{}).Where("id IN ?", ids).Update("gdelt_worker_id", worker.ID).Error
}

// GDELTSource stores BigQuery search parameters (see [GDELTSearchRequest]) plus optional natural-language gating fields.
type GDELTSource struct {
	gorm.Model

	GDELTWorkerID *uint        `gorm:"index"`
	GDELTWorker   *GDELTWorker `gorm:"foreignKey:GDELTWorkerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	Query         string
	Domain        string
	ActionCountry string
	StartDate     *time.Time
	EndDate       *time.Time
	MinMentions   uint   `gorm:"not null;default:0"`
	MaxRecords    uint   `gorm:"not null;default:0"`
	Sort          string `gorm:"size:32;not null;default:''"`

	// NaturalLanguageFilter is optional text (e.g. one rule per line) for whitelist/blacklist per [IsBlacklist], analogous to [p_seer_reddit.RedditSource.Filter] / [p_seer_reddit.RedditSource.IsFilterWhitelist].
	NaturalLanguageFilter string `gorm:"type:text;not null;default:''"`
	IsBlacklist           bool   `gorm:"not null;default:false"`
}

func (GDELTSource) TableName() string {
	return GDELTSourcesTable
}

// ToGDELTSearchRequest maps persisted fields to the BigQuery search shape used by [FetchAndStoreGDELTEvents].
// EndDate is normalized to end-of-day UTC to match [parseGDELTSearchRequest].
func (s *GDELTSource) ToGDELTSearchRequest() GDELTSearchRequest {
	req := GDELTSearchRequest{
		Query:         strings.TrimSpace(s.Query),
		Domain:        strings.TrimSpace(s.Domain),
		ActionCountry: strings.TrimSpace(s.ActionCountry),
		MinMentions:   s.MinMentions,
		MaxRecords:    s.MaxRecords,
		Sort:          strings.TrimSpace(s.Sort),
	}
	if s.StartDate != nil && !s.StartDate.IsZero() {
		t := s.StartDate.UTC()
		req.StartDate = &t
	}
	if s.EndDate != nil && !s.EndDate.IsZero() {
		t := s.EndDate.UTC()
		end := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
		req.EndDate = &end
	}
	return req
}

// Event stores one fetched GDELT event row using the daily updates schema, including SOURCEURL.
type Event struct {
	gorm.Model

	GDELTSourceID *uint        `gorm:"index"`
	GDELTSource   *GDELTSource `gorm:"foreignKey:GDELTSourceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	GlobalEventID uint64 `gorm:"not null;uniqueIndex"`
	SQLDate       int    `gorm:"index"`
	MonthYear     string `gorm:"size:6"`
	Year          string `gorm:"size:4"`
	FractionDate  float64

	Actor1Code           string `gorm:"size:32"`
	Actor1Name           string `gorm:"size:255"`
	Actor1CountryCode    string `gorm:"size:8"`
	Actor1KnownGroupCode string `gorm:"size:32"`
	Actor1EthnicCode     string `gorm:"size:32"`
	Actor1Religion1Code  string `gorm:"size:32"`
	Actor1Religion2Code  string `gorm:"size:32"`
	Actor1Type1Code      string `gorm:"size:32"`
	Actor1Type2Code      string `gorm:"size:32"`
	Actor1Type3Code      string `gorm:"size:32"`

	Actor2Code           string `gorm:"size:32"`
	Actor2Name           string `gorm:"size:255"`
	Actor2CountryCode    string `gorm:"size:8"`
	Actor2KnownGroupCode string `gorm:"size:32"`
	Actor2EthnicCode     string `gorm:"size:32"`
	Actor2Religion1Code  string `gorm:"size:32"`
	Actor2Religion2Code  string `gorm:"size:32"`
	Actor2Type1Code      string `gorm:"size:32"`
	Actor2Type2Code      string `gorm:"size:32"`
	Actor2Type3Code      string `gorm:"size:32"`

	IsRootEvent    int
	EventCode      string `gorm:"size:8"`
	EventBaseCode  string `gorm:"size:8"`
	EventRootCode  string `gorm:"size:8"`
	QuadClass      int
	GoldsteinScale float64
	NumMentions    int
	NumSources     int
	NumArticles    int
	AvgTone        float64

	Actor1GeoType        int
	Actor1GeoFullName    string `gorm:"size:255"`
	Actor1GeoCountryCode string `gorm:"size:8"`
	Actor1GeoADM1Code    string `gorm:"size:32"`
	Actor1GeoADM2Code    string `gorm:"size:32"`
	Actor1GeoLat         float64
	Actor1GeoLong        float64
	Actor1GeoFeatureID   string `gorm:"size:64"`

	Actor2GeoType        int
	Actor2GeoFullName    string `gorm:"size:255"`
	Actor2GeoCountryCode string `gorm:"size:8"`
	Actor2GeoADM1Code    string `gorm:"size:32"`
	Actor2GeoADM2Code    string `gorm:"size:32"`
	Actor2GeoLat         float64
	Actor2GeoLong        float64
	Actor2GeoFeatureID   string `gorm:"size:64"`

	ActionGeoType        int
	ActionGeoFullName    string       `gorm:"size:255"`
	ActionGeoCountryCode string       `gorm:"size:8"`
	ActionGeoADM1Code    string       `gorm:"size:32"`
	ActionGeoADM2Code    string       `gorm:"size:32"`
	ActionGeoPoint       lago.PGPoint `gorm:"type:point"`
	ActionGeoLat         float64      `gorm:"-"` // form roundtrip; persisted via [ActionGeoPoint]
	ActionGeoLong        float64      `gorm:"-"`
	ActionGeoFeatureID   string       `gorm:"size:64"`

	DateAdded int64
	SourceURL string `gorm:"size:1024"`
}

func (Event) TableName() string {
	return EventsTable
}

func (e *Event) AfterCreate(_ *gorm.DB) error {
	EnqueueEventSourceURLForWebsiteScrape(e.SourceURL)
	return nil
}

func (e *Event) AfterFind(_ *gorm.DB) error {
	e.syncActionGeoFloatsFromPoint()
	return nil
}

func (e *Event) BeforeSave(_ *gorm.DB) error {
	if gdeltValidLatLng(e.ActionGeoLat, e.ActionGeoLong) {
		e.ActionGeoPoint = lago.NewPGPoint(e.ActionGeoLong, e.ActionGeoLat)
	} else {
		e.ActionGeoPoint = lago.PGPoint{}
	}
	return nil
}

func (e *Event) syncActionGeoFloatsFromPoint() {
	if e.ActionGeoPoint.Valid {
		e.ActionGeoLat = e.ActionGeoPoint.P.Y
		e.ActionGeoLong = e.ActionGeoPoint.P.X
	} else {
		e.ActionGeoLat = 0
		e.ActionGeoLong = 0
	}
}

func init() {
	lago.OnDBInit("p_seer_gdelt.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[GDELTWorker](db)
		lago.RegisterModel[GDELTSource](db)
		lago.RegisterModel[Event](db)
		return db
	})
	lago.OnDBInit("p_seer_gdelt.worker_pools_autostart", func(db *gorm.DB) *gorm.DB {
		StartAllGDELTWorkerPools(db)
		return db
	})
}
