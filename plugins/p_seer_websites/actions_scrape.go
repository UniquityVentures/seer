package p_seer_websites

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/lago/views"
	"github.com/UniquityVentures/seer/plugins/p_seer_node_fleet"
	"github.com/UniquityVentures/seer/plugins/p_seer_node_fleet/messages"
)

var errSSRFAfterRedirect = errors.New("p_seer_websites: redirect landed on non-public host")

var fleetCommandID atomic.Uint64

func fleetScrapeError(resp *messages.Response) error {
	if resp == nil || resp.GetError() == nil {
		return fmt.Errorf("fleet scrape: unknown error")
	}
	if wsErr := resp.GetError().GetWebsiteScraperError(); wsErr != nil {
		if sf := wsErr.GetScrapeFailed(); sf != nil && strings.TrimSpace(sf.GetMessage()) != "" {
			return fmt.Errorf("fleet scrape failed: %s", sf.GetMessage())
		}
	}
	return fmt.Errorf("fleet scrape error: %v", resp.GetError())
}

func scrapeViaFleet(ctx context.Context, pageURL string) (markdown string, renderedHTML string, final *url.URL, err error) {
	cmdID := fleetCommandID.Add(1)
	cmd := &messages.Command{
		Id: cmdID,
		CommandType: &messages.Command_TriggerScraper{
			TriggerScraper: &messages.TriggerScraper{
				ScraperArgs: &messages.TriggerScraper_WebsiteScraper{
					WebsiteScraper: &messages.WebsiteScraperArgs{
						SourceUrl: pageURL,
					},
				},
			},
		},
	}
	resp, err := p_seer_node_fleet.DispatchCommand(cmd)
	if err != nil {
		return "", nil, fmt.Errorf("fleet dispatch: %w", err)
	}
	if resp.GetError() != nil {
		return "", nil, fleetScrapeError(resp)
	}
	ok := resp.GetOk()
	if ok == nil {
		return "", nil, fmt.Errorf("fleet scrape: empty ok response")
	}
	ws := ok.GetWebsiteScraper()
	if ws == nil {
		return "", nil, fmt.Errorf("fleet scrape: not a website scraper response")
	}
	if resp.GetCommandId() != 0 && resp.GetCommandId() != cmdID {
		return "", nil, fmt.Errorf("fleet scrape: response command id mismatch (sent %d, got %d)", cmdID, resp.GetCommandId())
	}
	html = ws.GetContent()
	finalU, err := url.Parse(ws.GetSourceUrl())
	if err != nil {
		return "", nil, err
	}
	slog.Info("p_seer_websites: fleet scrape returned HTML",
		"url", pageURL,
		"final_url", finalU.String(),
		"html_bytes", len(html),
	)
	if strings.TrimSpace(html) == "" {
		return "", nil, fmt.Errorf("fleet scrape returned empty HTML for %s (check nodescraper logs; Chrome/chromedriver must be available on the node)", pageURL)
	}
	return html, finalU, nil
}

// ScrapeToMarkdown validates URL, fetches rendered HTML via the node fleet, returns markdown and canonical URL.
func ScrapeToMarkdown(ctx context.Context, rawURL string) (markdown string, canonical *url.URL, err error) {
	canon, err := fetchableWebsiteURL(ctx, rawURL)
	if err != nil {
		return "", nil, err
	}
	htmlStr, finalU, err := scrapeHTMLViaFleet(ctx, canon.String())
	if err != nil {
		return "", nil, fmt.Errorf("fetch page: %w", err)
	}
	if urlFailsSSRF(ctx, finalU) {
		return "", nil, errSSRFAfterRedirect
	}
	md := markdownFromRenderedHTML(htmlStr, finalU)
	if strings.TrimSpace(md) == "" {
		return "", nil, fmt.Errorf("no extractable text from page (readability empty; fleet returned %d bytes HTML from %s)", len(htmlStr), finalU.String())
	}
	out := cloneURL(canon)
	if finalU != nil {
		if norm, e := normalizeWebsiteURL(finalU.String()); e == nil {
			out = cloneURL(norm)
		}
	}
	return md, out, nil
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	return new(*u)
}

// websiteScrapeFormPatcher fills [Website.Markdown] and normalizes [Website.URL] before [views.LayerCreate] persists.
type websiteScrapeFormPatcher struct{}

func (websiteScrapeFormPatcher) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if len(formErrors) > 0 {
		return formData, formErrors
	}
	raw, _ := formData["URL"].(string)
	md, canon, err := ScrapeToMarkdown(r.Context(), raw)
	if err != nil {
		formErrors["_form"] = err
		return formData, formErrors
	}
	var pp lago.PageURL
	pp.SetFromURL(canon)
	formData["URL"] = pp
	formData["Markdown"] = md
	return formData, formErrors
}

// websiteTitleHint returns a short title for intel rows from a page URL.
func websiteTitleHint(u *url.URL) string {
	if u == nil || u.Host == "" {
		return "Website"
	}
	return strings.TrimPrefix(u.Hostname(), "www.")
}
