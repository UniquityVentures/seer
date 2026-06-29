package p_seer_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/UniquityVentures/lamu/plugins/p_google_genai"
	"github.com/UniquityVentures/lamu/plugins/p_llm_assistant"
	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
	"github.com/UniquityVentures/seer/plugins/p_seer_reddit"
	"github.com/UniquityVentures/seer/plugins/p_seer_websites"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

// -- Argument Structs --

type intelSearchArgs struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type redditAddSourceArgs struct {
	RedditRunnerID    *uint    `json:"reddit_runner_id,omitempty"`
	Subreddits        []string `json:"subreddits,omitempty"`
	SearchQuery       string   `json:"search_query,omitempty"`
	Filter            string   `json:"filter,omitempty"`
	IsFilterWhitelist bool     `json:"is_filter_whitelist,omitempty"`
	MaxFreshPosts     uint     `json:"max_fresh_posts,omitempty"`
	LoadWebsites      bool     `json:"load_websites,omitempty"`
}

type redditEditSourceArgs struct {
	RedditSourceID    uint     `json:"reddit_source_id"`
	RedditRunnerID    *uint    `json:"reddit_runner_id,omitempty"`
	Subreddits        []string `json:"subreddits,omitempty"`
	SearchQuery       string   `json:"search_query,omitempty"`
	Filter            string   `json:"filter,omitempty"`
	IsFilterWhitelist bool     `json:"is_filter_whitelist,omitempty"`
	MaxFreshPosts     uint     `json:"max_fresh_posts,omitempty"`
	LoadWebsites      bool     `json:"load_websites,omitempty"`
}

type redditAddWorkerArgs struct {
	WorkerName     string `json:"worker_name"`
	WorkerDuration string `json:"worker_duration"`
}

type redditEditWorkerArgs struct {
	RedditRunnerID *uint  `json:"reddit_runner_id"`
	WorkerName     string `json:"worker_name"`
	WorkerDuration string `json:"worker_duration"`
}

type websiteListArgs struct{}

type websiteAddSourceArgs struct {
	SeedURL         string `json:"seed_url"`
	WebsiteDepth    uint   `json:"website_depth,omitempty"`
	WebsiteRunnerFK *uint  `json:"website_runner_id,omitempty"`
}

type websiteEditSourceArgs struct {
	WebsiteSourceID uint   `json:"website_source_id"`
	SeedURL         string `json:"seed_url"`
	WebsiteDepth    uint   `json:"website_depth,omitempty"`
	WebsiteRunnerFK *uint  `json:"website_runner_id,omitempty"`
}

type websiteAddWorkerArgs struct {
	WorkerName     string `json:"worker_name"`
	WorkerDuration string `json:"worker_duration"`
}

type websiteEditWorkerArgs struct {
	WebsiteRunnerFK uint   `json:"website_runner_id"`
	WorkerName      string `json:"worker_name"`
	WorkerDuration  string `json:"worker_duration"`
}

// -- Helper to convert string results to maps --

func toMapResult(resText, errText string) (map[string]any, error) {
	if strings.TrimSpace(errText) != "" {
		return nil, fmt.Errorf("%s", errText)
	}
	if strings.TrimSpace(resText) == "" {
		return map[string]any{"output": ""}, nil
	}
	var parsed any
	if err := json.Unmarshal([]byte(resText), &parsed); err == nil {
		if m, ok := parsed.(map[string]any); ok {
			return m, nil
		}
		return map[string]any{"output": parsed}, nil
	}
	return map[string]any{"output": resText}, nil
}

// -- LlmTool Implementations --

// Intel Search Tool
type intelSearchTool struct{}

func (t *intelSearchTool) Name() string { return "intel_search" }
func (t *intelSearchTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "intel_search",
		Description: "Vector search over the Intel database (summaries of scraped pages and ingested content). Returns id, title, summary, kind.",
		Parameters:  p_google_genai.NewSchema[intelSearchArgs](),
	}
}
func (t *intelSearchTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	var a intelSearchArgs
	if b, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	res, errText := runIntelSearchTool(ctx, db, a.Query, a.Limit)
	return toMapResult(res, errText)
}

// Reddit Add Source Tool
type redditAddSourceTool struct{}

func (t *redditAddSourceTool) Name() string { return "reddit_add_source" }
func (t *redditAddSourceTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name: "reddit_add_source",
		Description: "Create a Reddit ingestion source (subreddits and/or search query, optional runner). " +
			"Optional filter: non-empty filter triggers an LLM gate on ingest; is_filter_whitelist true means only matching posts are kept, false means matching posts are rejected (blacklist).",
		Parameters: p_google_genai.NewSchema[redditAddSourceArgs](),
	}
}
func (t *redditAddSourceTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	var a redditAddSourceArgs
	if b, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	res, errText := runRedditAddSourceTool(ctx, db, a)
	return toMapResult(res, errText)
}

// Reddit Edit Source Tool
type redditEditSourceTool struct{}

func (t *redditEditSourceTool) Name() string { return "reddit_edit_source" }
func (t *redditEditSourceTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "reddit_edit_source",
		Description: "Update an existing Reddit source by reddit_source_id (same fields as create, including filter and is_filter_whitelist).",
		Parameters:  p_google_genai.NewSchema[redditEditSourceArgs](),
	}
}
func (t *redditEditSourceTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	var a redditEditSourceArgs
	if b, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	res, errText := runRedditEditSourceTool(ctx, db, a)
	return toMapResult(res, errText)
}

// Reddit Add Worker Tool
type redditAddWorkerTool struct{}

func (t *redditAddWorkerTool) Name() string { return "reddit_add_worker" }
func (t *redditAddWorkerTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "reddit_add_worker",
		Description: "Create a Reddit runner worker schedule (worker_name, worker_duration as Go duration string e.g. 1h, 30m).",
		Parameters:  p_google_genai.NewSchema[redditAddWorkerArgs](),
	}
}
func (t *redditAddWorkerTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	var a redditAddWorkerArgs
	if b, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	res, errText := runRedditAddWorkerTool(ctx, db, a)
	return toMapResult(res, errText)
}

// Reddit Edit Worker Tool
type redditEditWorkerTool struct{}

func (t *redditEditWorkerTool) Name() string { return "reddit_edit_worker" }
func (t *redditEditWorkerTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "reddit_edit_worker",
		Description: "Update a Reddit runner worker schedule (reddit_runner_id, worker name, Go duration string e.g. 1h, 30m).",
		Parameters:  p_google_genai.NewSchema[redditEditWorkerArgs](),
	}
}
func (t *redditEditWorkerTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	var a redditEditWorkerArgs
	if b, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	res, errText := runRedditEditWorkerTool(ctx, db, a)
	return toMapResult(res, errText)
}

// Website List Sources Tool
type websiteListSourcesTool struct{}

func (t *websiteListSourcesTool) Name() string { return "website_list_sources" }
func (t *websiteListSourcesTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "website_list_sources",
		Description: "List Seer Websites crawl sources (JSON).",
		Parameters:  p_google_genai.NewSchema[websiteListArgs](),
	}
}
func (t *websiteListSourcesTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	res, errText := runWebsiteListSourcesTool(ctx, db)
	return toMapResult(res, errText)
}

// Website List Workers Tool
type websiteListWorkersTool struct{}

func (t *websiteListWorkersTool) Name() string { return "website_list_workers" }
func (t *websiteListWorkersTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "website_list_workers",
		Description: "List Seer Websites crawl workers/runners (JSON).",
		Parameters:  p_google_genai.NewSchema[websiteListArgs](),
	}
}
func (t *websiteListWorkersTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	res, errText := runWebsiteListWorkersTool(ctx, db)
	return toMapResult(res, errText)
}

// Website Add Source Tool
type websiteAddSourceTool struct{}

func (t *websiteAddSourceTool) Name() string { return "website_add_source" }
func (t *websiteAddSourceTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "website_add_source",
		Description: "Add a website crawl source (seed_url, optional depth and website_runner_id).",
		Parameters:  p_google_genai.NewSchema[websiteAddSourceArgs](),
	}
}
func (t *websiteAddSourceTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	var a websiteAddSourceArgs
	if b, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	res, errText := runWebsiteAddSourceTool(ctx, db, a)
	return toMapResult(res, errText)
}

// Website Edit Source Tool
type websiteEditSourceTool struct{}

func (t *websiteEditSourceTool) Name() string { return "website_edit_source" }
func (t *websiteEditSourceTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "website_edit_source",
		Description: "Edit a website source by website_source_id (seed_url, optional depth and runner).",
		Parameters:  p_google_genai.NewSchema[websiteEditSourceArgs](),
	}
}
func (t *websiteEditSourceTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	var a websiteEditSourceArgs
	if b, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	res, errText := runWebsiteEditSourceTool(ctx, db, a)
	return toMapResult(res, errText)
}

// Website Add Worker Tool
type websiteAddWorkerTool struct{}

func (t *websiteAddWorkerTool) Name() string { return "website_add_worker" }
func (t *websiteAddWorkerTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "website_add_worker",
		Description: "Add a website crawl worker (worker_name, worker_duration as Go duration string).",
		Parameters:  p_google_genai.NewSchema[websiteAddWorkerArgs](),
	}
}
func (t *websiteAddWorkerTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	var a websiteAddWorkerArgs
	if b, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	res, errText := runWebsiteAddWorkerTool(ctx, db, a)
	return toMapResult(res, errText)
}

// Website Edit Worker Tool
type websiteEditWorkerTool struct{}

func (t *websiteEditWorkerTool) Name() string { return "website_edit_worker" }
func (t *websiteEditWorkerTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "website_edit_worker",
		Description: "Edit a website worker by website_runner_id (name and duration).",
		Parameters:  p_google_genai.NewSchema[websiteEditWorkerArgs](),
	}
}
func (t *websiteEditWorkerTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	var a websiteEditWorkerArgs
	if b, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	res, errText := runWebsiteEditWorkerTool(ctx, db, a)
	return toMapResult(res, errText)
}

// -- init registration --

func init() {
	p_llm_assistant.LlmToolRegistry.Register("intel_search", &intelSearchTool{})
	p_llm_assistant.LlmToolRegistry.Register("reddit_add_source", &redditAddSourceTool{})
	p_llm_assistant.LlmToolRegistry.Register("reddit_edit_source", &redditEditSourceTool{})
	p_llm_assistant.LlmToolRegistry.Register("reddit_add_worker", &redditAddWorkerTool{})
	p_llm_assistant.LlmToolRegistry.Register("reddit_edit_worker", &redditEditWorkerTool{})
	p_llm_assistant.LlmToolRegistry.Register("website_list_sources", &websiteListSourcesTool{})
	p_llm_assistant.LlmToolRegistry.Register("website_list_workers", &websiteListWorkersTool{})
	p_llm_assistant.LlmToolRegistry.Register("website_add_source", &websiteAddSourceTool{})
	p_llm_assistant.LlmToolRegistry.Register("website_edit_source", &websiteEditSourceTool{})
	p_llm_assistant.LlmToolRegistry.Register("website_add_worker", &websiteAddWorkerTool{})
	p_llm_assistant.LlmToolRegistry.Register("website_edit_worker", &websiteEditWorkerTool{})
}

// -- Underlying tool logic functions --

func runIntelSearchTool(ctx context.Context, db *gorm.DB, query string, limit int) (string, string) {
	q := strings.TrimSpace(query)
	if q == "" {
		return "", "intel_search: empty query"
	}
	if limit <= 0 {
		limit = 8
	}
	if limit > 20 {
		limit = 20
	}
	rows, err := p_seer_intel.SearchIntelBySimilarity(ctx, db, q, limit)
	if err != nil {
		return "", err.Error()
	}
	if len(rows) == 0 {
		return "[]", ""
	}
	type row struct {
		ID      uint   `json:"id"`
		Title   string `json:"title"`
		Summary string `json:"summary"`
		Kind    string `json:"kind"`
		KindID  uint   `json:"kind_id"`
	}
	out := make([]row, 0, len(rows))
	for _, r := range rows {
		out = append(out, row{ID: r.ID, Title: r.Title, Summary: r.Summary, Kind: r.Kind, KindID: r.KindID})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}

func runRedditAddSourceTool(ctx context.Context, tx *gorm.DB, a redditAddSourceArgs) (string, string) {
	var runner *uint
	if a.RedditRunnerID != nil && *a.RedditRunnerID != 0 {
		id := *a.RedditRunnerID
		runner = &id
	}
	p := p_seer_reddit.RedditSourceCreateParams{
		RedditRunnerID:    runner,
		Subreddits:        a.Subreddits,
		SearchQuery:       a.SearchQuery,
		Filter:            a.Filter,
		IsFilterWhitelist: a.IsFilterWhitelist,
		MaxFreshPosts:     a.MaxFreshPosts,
		LoadWebsites:      a.LoadWebsites,
	}
	src, err := p_seer_reddit.CreateRedditSource(ctx, tx, p)
	if err != nil {
		return "", err.Error()
	}
	return fmt.Sprintf(`{"reddit_source_id":%d}`, src.ID), ""
}

func runRedditEditSourceTool(ctx context.Context, tx *gorm.DB, a redditEditSourceArgs) (string, string) {
	if a.RedditSourceID == 0 {
		return "", "reddit_edit_source: reddit_source_id required"
	}
	var runner *uint
	if a.RedditRunnerID != nil && *a.RedditRunnerID != 0 {
		id := *a.RedditRunnerID
		runner = &id
	}
	p := p_seer_reddit.RedditSourceUpdateParams{
		SourceID: a.RedditSourceID,
		RedditSourceCreateParams: p_seer_reddit.RedditSourceCreateParams{
			RedditRunnerID:    runner,
			Subreddits:        a.Subreddits,
			SearchQuery:       a.SearchQuery,
			Filter:            a.Filter,
			IsFilterWhitelist: a.IsFilterWhitelist,
			MaxFreshPosts:     a.MaxFreshPosts,
			LoadWebsites:      a.LoadWebsites,
		},
	}
	src, err := p_seer_reddit.UpdateRedditSource(ctx, tx, p)
	if err != nil {
		return "", err.Error()
	}
	return fmt.Sprintf(`{"reddit_source_id":%d}`, src.ID), ""
}

func runRedditAddWorkerTool(ctx context.Context, tx *gorm.DB, a redditAddWorkerArgs) (string, string) {
	name := strings.TrimSpace(a.WorkerName)
	if name == "" {
		return "", "reddit_add_worker: worker_name required"
	}
	durStr := strings.TrimSpace(a.WorkerDuration)
	if durStr == "" {
		return "", "reddit_add_worker: worker_duration required (Go duration, e.g. 1h, 45m, 90s)"
	}
	d, err := time.ParseDuration(durStr)
	if err != nil {
		return "", fmt.Sprintf("reddit_add_worker: invalid worker_duration: %v", err)
	}
	runner, err := p_seer_reddit.CreateRedditRunner(ctx, tx, p_seer_reddit.RedditRunnerCreateParams{
		Name:     name,
		Duration: d,
	})
	if err != nil {
		return "", err.Error()
	}
	out := map[string]any{
		"reddit_runner_id": runner.ID,
		"name":             runner.Name,
		"duration":         runner.Duration.String(),
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}

func runRedditEditWorkerTool(ctx context.Context, tx *gorm.DB, a redditEditWorkerArgs) (string, string) {
	var rid uint
	if a.RedditRunnerID != nil && *a.RedditRunnerID != 0 {
		rid = *a.RedditRunnerID
	}
	if rid == 0 {
		return "", "reddit_edit_worker: reddit_runner_id required"
	}
	name := strings.TrimSpace(a.WorkerName)
	if name == "" {
		return "", "reddit_edit_worker: worker_name required"
	}
	durStr := strings.TrimSpace(a.WorkerDuration)
	if durStr == "" {
		return "", "reddit_edit_worker: worker_duration required (Go duration, e.g. 1h, 45m, 90s)"
	}
	d, err := time.ParseDuration(durStr)
	if err != nil {
		return "", fmt.Sprintf("reddit_edit_worker: invalid worker_duration: %v", err)
	}
	runner, err := p_seer_reddit.UpdateRedditRunner(ctx, tx, p_seer_reddit.RedditRunnerUpdateParams{
		ID:       rid,
		Name:     name,
		Duration: d,
	})
	if err != nil {
		return "", err.Error()
	}
	out := map[string]any{
		"reddit_runner_id": runner.ID,
		"name":             runner.Name,
		"duration":         runner.Duration.String(),
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}

func runWebsiteListSourcesTool(ctx context.Context, db *gorm.DB) (string, string) {
	s, err := p_seer_websites.ListWebsiteSourcesJSON(ctx, db)
	if err != nil {
		return "", err.Error()
	}
	return s, ""
}

func runWebsiteListWorkersTool(ctx context.Context, db *gorm.DB) (string, string) {
	s, err := p_seer_websites.ListWebsiteRunnersJSON(ctx, db)
	if err != nil {
		return "", err.Error()
	}
	return s, ""
}

func runWebsiteAddSourceTool(ctx context.Context, tx *gorm.DB, a websiteAddSourceArgs) (string, string) {
	seed := strings.TrimSpace(a.SeedURL)
	if seed == "" {
		return "", "website_add_source: seed_url required"
	}
	src, err := p_seer_websites.CreateWebsiteSourceFromParams(ctx, tx, p_seer_websites.WebsiteSourceCreateParams{
		SeedURL:         seed,
		Depth:           a.WebsiteDepth,
		WebsiteRunnerID: a.WebsiteRunnerFK,
	})
	if err != nil {
		return "", err.Error()
	}
	out := map[string]any{
		"website_source_id": src.ID,
		"seed_url":          strings.TrimSpace(src.URL.String()),
		"website_depth":     src.Depth,
	}
	if src.WebsiteRunnerID != nil {
		out["website_runner_id"] = *src.WebsiteRunnerID
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}

func runWebsiteEditSourceTool(ctx context.Context, tx *gorm.DB, a websiteEditSourceArgs) (string, string) {
	if a.WebsiteSourceID == 0 {
		return "", "website_edit_source: website_source_id required"
	}
	seed := strings.TrimSpace(a.SeedURL)
	if seed == "" {
		return "", "website_edit_source: seed_url required"
	}
	src, err := p_seer_websites.UpdateWebsiteSourceFromParams(ctx, tx, p_seer_websites.WebsiteSourceUpdateParams{
		SourceID:        a.WebsiteSourceID,
		SeedURL:         seed,
		Depth:           a.WebsiteDepth,
		WebsiteRunnerID: a.WebsiteRunnerFK,
	})
	if err != nil {
		return "", err.Error()
	}
	out := map[string]any{
		"website_source_id": src.ID,
		"seed_url":          strings.TrimSpace(src.URL.String()),
		"website_depth":     src.Depth,
	}
	if src.WebsiteRunnerID != nil {
		out["website_runner_id"] = *src.WebsiteRunnerID
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}

func runWebsiteAddWorkerTool(ctx context.Context, tx *gorm.DB, a websiteAddWorkerArgs) (string, string) {
	name := strings.TrimSpace(a.WorkerName)
	durStr := strings.TrimSpace(a.WorkerDuration)
	if name == "" {
		return "", "website_add_worker: worker_name required"
	}
	if durStr == "" {
		return "", "website_add_worker: worker_duration required (Go duration, e.g. 1h, 30m)"
	}
	d, err := time.ParseDuration(durStr)
	if err != nil {
		return "", fmt.Sprintf("website_add_worker: invalid worker_duration: %v", err)
	}
	runner, err := p_seer_websites.CreateWebsiteRunnerFromParams(ctx, tx, p_seer_websites.WebsiteRunnerCreateParams{
		Name:     name,
		Duration: d,
	})
	if err != nil {
		return "", err.Error()
	}
	out := map[string]any{
		"website_runner_id": runner.ID,
		"name":              runner.Name,
		"duration":          runner.Duration.String(),
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}

func runWebsiteEditWorkerTool(ctx context.Context, tx *gorm.DB, a websiteEditWorkerArgs) (string, string) {
	if a.WebsiteRunnerFK == 0 {
		return "", "website_edit_worker: website_runner_id required"
	}
	name := strings.TrimSpace(a.WorkerName)
	durStr := strings.TrimSpace(a.WorkerDuration)
	if name == "" {
		return "", "website_edit_worker: worker_name required"
	}
	if durStr == "" {
		return "", "website_edit_worker: worker_duration required"
	}
	d, err := time.ParseDuration(durStr)
	if err != nil {
		return "", fmt.Sprintf("website_edit_worker: invalid worker_duration: %v", err)
	}
	runner, err := p_seer_websites.UpdateWebsiteRunnerFromParams(ctx, tx, p_seer_websites.WebsiteRunnerUpdateParams{
		ID:       a.WebsiteRunnerFK,
		Name:     name,
		Duration: d,
	})
	if err != nil {
		return "", err.Error()
	}
	out := map[string]any{
		"website_runner_id": runner.ID,
		"name":              runner.Name,
		"duration":          runner.Duration.String(),
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}
