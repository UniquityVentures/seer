package p_seer_reddit

import (
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/UniquityVentures/seer/plugins/p_seer_websites"
)

var redditPostHTTPURLRe = regexp.MustCompile(`https?://[^\s<>()\[\]"']+`)

func websiteScrapeHostSkipped(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return true
	}
	if h == "reddit.com" || strings.HasSuffix(h, ".reddit.com") {
		return true
	}
	return false
}

func tryEnqueueWebsiteScrapeURL(raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return
	}
	if websiteScrapeHostSkipped(parsed.Hostname()) {
		return
	}
	toSend, err := url.Parse(parsed.String())
	if err != nil || toSend.Host == "" {
		return
	}
	p_seer_websites.WebsiteScrapeURLQueue <- toSend
}

// allHTTPURLsInHTML returns every http(s) URL found in HTML (valid parse, deduped, document order).
func allHTTPURLsInHTML(htmlSrc string) []string {
	s := html.UnescapeString(htmlSrc)
	seen := make(map[string]struct{})
	var out []string
	for _, m := range redditPostHTTPURLRe.FindAllString(s, -1) {
		u := strings.TrimRight(m, ".,;:!?)")
		parsed, err := url.Parse(u)
		if err != nil || parsed.Host == "" {
			continue
		}
		switch strings.ToLower(parsed.Scheme) {
		case "http", "https":
		default:
			continue
		}
		if _, dup := seen[u]; dup {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

// firstOutboundHTTPURLInHTML returns the first URL whose host is not a Reddit scrape-skip host (outbound link for stored post URL).
func firstOutboundHTTPURLInHTML(htmlSrc string) string {
	for _, u := range allHTTPURLsInHTML(htmlSrc) {
		parsed, err := url.Parse(u)
		if err != nil || parsed.Host == "" {
			continue
		}
		if !websiteScrapeHostSkipped(parsed.Hostname()) {
			return u
		}
	}
	return ""
}

// enqueueWebsiteURLFromRSSContent enqueues every http(s) URL from Reddit Atom entry content; tryEnqueueWebsiteScrapeURL skips unsuitable hosts.
func enqueueWebsiteURLFromRSSContent(html string) {
	for _, u := range allHTTPURLsInHTML(html) {
		tryEnqueueWebsiteScrapeURL(u)
	}
}
