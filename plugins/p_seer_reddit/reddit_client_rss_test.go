package p_seer_reddit

import (
	"encoding/xml"
	"testing"
)

func TestParseRedditAtomFeedSample(t *testing.T) {
	const sample = `<?xml version="1.0" encoding="UTF-8"?><feed xmlns="http://www.w3.org/2005/Atom"><entry><id>t3_abc123</id><title>Hello</title><updated>2026-04-30T19:01:41+00:00</updated><published>2026-04-30T19:01:41+00:00</published><author><name>/u/x</name></author><category term="golang" label="r/golang"/><content type="html">&lt;p&gt;Hi &lt;a href=&quot;https://example.com/z&quot;&gt;link&lt;/a&gt;&lt;/p&gt;</content><link href="https://www.reddit.com/r/golang/comments/abc123/hello/" /></entry></feed>`
	var feed redditAtomFeed
	if err := xml.Unmarshal([]byte(sample), &feed); err != nil {
		t.Fatal(err)
	}
	if len(feed.Entry) != 1 {
		t.Fatalf("entries: %d", len(feed.Entry))
	}
	if feed.Entry[0].Content.Body == "" {
		t.Fatal("expected content body populated")
	}
	post, err := redditPostDataFromAtomEntry("golang", feed.Entry[0])
	if err != nil {
		t.Fatal(err)
	}
	if post.ID != "abc123" {
		t.Fatalf("id %q", post.ID)
	}
	if post.URL != "https://example.com/z" {
		t.Fatalf("url %q", post.URL)
	}
	if post.Permalink != "/r/golang/comments/abc123/hello/" {
		t.Fatalf("permalink %q", post.Permalink)
	}
}
