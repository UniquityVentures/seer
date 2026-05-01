package p_seer_reddit

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RedditPostData holds post fields used for persistence, LLM filter, and RSS-derived website URL extraction.
type RedditPostData struct {
	ID             string
	Name           string
	Subreddit      string
	SubredditID    string
	SubredditType  string
	Author         string
	AuthorFullname string
	Title          string
	Selftext       string
	SelftextHTML   string
	URL            string
	Permalink      string
	CreatedUTC     float64
	Edited         any
	Score          int
	Ups            int
	Downs          int
	NumComments    int
	IsSelf         bool
	IsVideo        bool
	Domain         string
	Thumbnail      string
	Over18         bool
	Spoiler        bool
	Stickied       bool
	Locked         bool
	Archived       bool
	RemovedBy      *string
	// AtomContentHTML is raw entry &lt;content&gt; inner HTML from Reddit Atom (used for website scrape URL extraction).
	AtomContentHTML string `json:"-"`
}

const redditRSSFetchLimit = 100

var htmlTagStripRe = regexp.MustCompile(`(?i)<[^>]+>`)

func redditUserAgent() string {
	return "lago:p_seer_reddit:1.0 (by /u/local)"
}

// redditSubPathSegment escapes a subreddit path segment; preserves "+" for multireddits.
func redditSubPathSegment(sub string) string {
	sub = strings.TrimSpace(strings.ReplaceAll(sub, "/", ""))
	return strings.ReplaceAll(url.PathEscape(sub), "%2B", "+")
}

func redditRSSNewFeedURL(subreddit string) string {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(redditRSSFetchLimit))
	return fmt.Sprintf("https://www.reddit.com/r/%s/new/.rss?%s", redditSubPathSegment(subreddit), q.Encode())
}

func redditRSSSearchFeedURL(subreddit, searchQuery string) string {
	q := url.Values{}
	q.Set("q", searchQuery)
	q.Set("restrict_sr", "1")
	q.Set("sort", "new")
	q.Set("limit", strconv.Itoa(redditRSSFetchLimit))
	return fmt.Sprintf("https://www.reddit.com/r/%s/search.rss?%s", redditSubPathSegment(subreddit), q.Encode())
}

func redditHTTPGet(ctx context.Context, fullURL string, logLabel string, subreddit string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		slog.Error("p_seer_reddit: new request", "error", err, "label", logLabel, "subreddit", subreddit)
		return nil, err
	}
	req.Header.Set("User-Agent", redditUserAgent())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("p_seer_reddit: http", "error", err, "label", logLabel, "subreddit", subreddit)
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		slog.Error("p_seer_reddit: read body", "error", err, "label", logLabel, "subreddit", subreddit)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("reddit: %s", resp.Status)
		slog.Error("p_seer_reddit: bad status", "error", err, "label", logLabel, "subreddit", subreddit)
		return nil, err
	}
	return b, nil
}

type redditAtomFeed struct {
	XMLName xml.Name          `xml:"http://www.w3.org/2005/Atom feed"`
	Entry   []redditAtomEntry `xml:"http://www.w3.org/2005/Atom entry"`
}

type redditAtomEntry struct {
	ID        string           `xml:"http://www.w3.org/2005/Atom id"`
	Title     string           `xml:"http://www.w3.org/2005/Atom title"`
	Updated   string           `xml:"http://www.w3.org/2005/Atom updated"`
	Published string           `xml:"http://www.w3.org/2005/Atom published"`
	Author    redditAtomAuthor `xml:"http://www.w3.org/2005/Atom author"`
	Content   redditAtomText   `xml:"http://www.w3.org/2005/Atom content"`
	Link      []redditAtomLink `xml:"http://www.w3.org/2005/Atom link"`
	Category  []redditAtomCat  `xml:"http://www.w3.org/2005/Atom category"`
}

type redditAtomAuthor struct {
	Name string `xml:"http://www.w3.org/2005/Atom name"`
}

type redditAtomText struct {
	Type string `xml:"type,attr"`
	Body string `xml:",chardata"`
}

type redditAtomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type redditAtomCat struct {
	Term string `xml:"term,attr"`
}

func parseRedditAtomFeed(body []byte) (*redditAtomFeed, error) {
	var feed redditAtomFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}
	return &feed, nil
}

func stripHTMLToSelftext(s string, maxRunes int) string {
	s = html.UnescapeString(s)
	s = htmlTagStripRe.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	if maxRunes > 0 && len([]rune(s)) > maxRunes {
		r := []rune(s)
		s = string(r[:maxRunes])
	}
	return strings.TrimSpace(s)
}

func atomEntryPostID(id string) string {
	id = strings.TrimSpace(id)
	id = strings.TrimPrefix(id, "tag:reddit.com:")
	if i := strings.LastIndex(id, ":"); i >= 0 && i+1 < len(id) {
		id = id[i+1:]
	}
	id = strings.TrimPrefix(id, "t3_")
	return id
}

func atomEntryPermalink(links []redditAtomLink) string {
	for _, ln := range links {
		h := strings.TrimSpace(ln.Href)
		if h == "" {
			continue
		}
		if strings.Contains(h, "/comments/") {
			u, err := url.Parse(h)
			if err != nil {
				continue
			}
			if u.Path != "" {
				return u.Path
			}
		}
	}
	return ""
}

func atomTimeUnix(updated, published string) (float64, error) {
	for _, raw := range []string{published, updated} {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			continue
		}
		return float64(t.Unix()), nil
	}
	return 0, fmt.Errorf("no parseable time")
}

func atomAuthorName(a redditAtomAuthor) string {
	n := strings.TrimSpace(a.Name)
	n = strings.TrimPrefix(n, "/u/")
	return n
}

func atomSubredditName(subredditFromFetch string, cats []redditAtomCat) string {
	for _, c := range cats {
		if t := strings.TrimSpace(c.Term); t != "" {
			return t
		}
	}
	return strings.TrimSpace(subredditFromFetch)
}

// redditPostDataFromAtomEntry maps one Atom entry into RedditPostData (subredditName is the listing’s subreddit).
func redditPostDataFromAtomEntry(subredditName string, e redditAtomEntry) (RedditPostData, error) {
	pid := atomEntryPostID(e.ID)
	if pid == "" {
		return RedditPostData{}, fmt.Errorf("empty post id")
	}
	created, err := atomTimeUnix(e.Updated, e.Published)
	if err != nil {
		return RedditPostData{}, err
	}
	contentHTML := strings.TrimSpace(e.Content.Body)
	extURL := firstOutboundHTTPURLInHTML(contentHTML)
	selftext := stripHTMLToSelftext(contentHTML, 50000)
	sub := atomSubredditName(subredditName, e.Category)
	permalink := atomEntryPermalink(e.Link)
	if permalink == "" {
		return RedditPostData{}, fmt.Errorf("missing permalink link")
	}
	domain := ""
	if extURL != "" {
		if u, err := url.Parse(extURL); err == nil && u.Host != "" {
			domain = u.Hostname()
		}
	}
	return RedditPostData{
		ID:              pid,
		Name:            "t3_" + pid,
		Subreddit:       sub,
		Author:          atomAuthorName(e.Author),
		Title:           strings.TrimSpace(e.Title),
		Selftext:        selftext,
		URL:             extURL,
		Permalink:       permalink,
		CreatedUTC:      created,
		Score:           0,
		NumComments:     0,
		IsSelf:          extURL == "",
		AtomContentHTML: contentHTML,
		Domain:          domain,
	}, nil
}

// fetchSubredditRSS returns posts from Reddit’s public Atom feed for r/{sub}/new.
func fetchSubredditRSS(ctx context.Context, subreddit string) ([]RedditPostData, error) {
	u := redditRSSNewFeedURL(subreddit)
	body, err := redditHTTPGet(ctx, u, "rss_new", subreddit)
	if err != nil {
		return nil, err
	}
	feed, err := parseRedditAtomFeed(body)
	if err != nil {
		slog.Error("p_seer_reddit: decode atom feed", "error", err, "subreddit", subreddit)
		return nil, err
	}
	var out []RedditPostData
	for _, ent := range feed.Entry {
		post, err := redditPostDataFromAtomEntry(subreddit, ent)
		if err != nil {
			slog.Warn("p_seer_reddit: skip atom entry", "error", err, "subreddit", subreddit)
			continue
		}
		out = append(out, post)
	}
	return out, nil
}

// fetchSubredditRSSSearch returns posts from Reddit’s search Atom feed restricted to the subreddit.
func fetchSubredditRSSSearch(ctx context.Context, subreddit, query string) ([]RedditPostData, error) {
	u := redditRSSSearchFeedURL(subreddit, query)
	body, err := redditHTTPGet(ctx, u, "rss_search", subreddit)
	if err != nil {
		return nil, err
	}
	feed, err := parseRedditAtomFeed(body)
	if err != nil {
		slog.Error("p_seer_reddit: decode atom search", "error", err, "subreddit", subreddit)
		return nil, err
	}
	var out []RedditPostData
	for _, ent := range feed.Entry {
		post, err := redditPostDataFromAtomEntry(subreddit, ent)
		if err != nil {
			slog.Warn("p_seer_reddit: skip atom entry (search)", "error", err, "subreddit", subreddit)
			continue
		}
		out = append(out, post)
	}
	return out, nil
}
