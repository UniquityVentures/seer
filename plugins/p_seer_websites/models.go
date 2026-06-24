package p_seer_websites

import (
	"time"

	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
	"gorm.io/gorm"
)

const WebsitesTable = "seer_websites"

const (
	WebsiteRunnersTable = "seer_website_runners"
	WebsiteSourcesTable = "seer_website_sources"
)

// Website stores a scraped page as markdown plus the canonical URL.
type Website struct {
	gorm.Model

	URL      lamu.PageURL `gorm:"column:url;type:text;not null;default:''"`
	Markdown string       `gorm:"type:text;not null;default:''"`
}

func (Website) TableName() string {
	return WebsitesTable
}

func (w *Website) AfterCreate(tx *gorm.DB) error {
	p_seer_intel.IntelChannel <- p_seer_intel.IngestRequest{
		Kind: w,
	}
	return nil
}

// WebsiteRunner is a cadence bucket for scheduled website source crawls ([WebsiteSource.WebsiteRunnerID] optional).
type WebsiteRunner struct {
	gorm.Model

	Name     string        `gorm:"size:64;not null;uniqueIndex"`
	Duration time.Duration `gorm:"not null"`
}

func (WebsiteRunner) TableName() string {
	return WebsiteRunnersTable
}

// WebsiteSource configures a seed URL, crawl depth, and optional worker runner.
//
// Depth is extra link hops after the seed page: 0 = seed only, 1 = seed plus direct links, etc.
// Discovered URLs must share the same origin (scheme + host + port) as the seed URL after its first successful navigation.
type WebsiteSource struct {
	gorm.Model

	WebsiteRunnerID *uint          `gorm:"index"`
	WebsiteRunner   *WebsiteRunner `gorm:"foreignKey:WebsiteRunnerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	URL   lamu.PageURL `gorm:"column:url;type:text;not null;default:''"`
	Depth uint         `gorm:"not null;default:0"`
}

func (WebsiteSource) TableName() string {
	return WebsiteSourcesTable
}

func init() {
	registerPluginDBInitHook("p_seer_websites.models", func(db *gorm.DB) *gorm.DB {
		if err := db.AutoMigrate(&Website{}, &WebsiteRunner{}, &WebsiteSource{}); err != nil {
			panic(err)
		}
		return db
	})
	registerPluginDBInitHook("p_seer_websites.runner_pools_autostart", func(db *gorm.DB) *gorm.DB {
		StartAllWebsiteRunnerWorkerPools(db)
		return db
	})
}
