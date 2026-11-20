package sourceadapter

import (
	"strings"
	"testing"
)

func TestExtractFeedCandidatesRSSAndAtom(t *testing.T) {
	rss := `<rss><channel>
      <item><title>Reforma de baño en Valencia</title><link>https://example.com/obras/1</link><description>Albañilería y pintura</description><pubDate>Fri, 20 Nov 2026 10:00:00 GMT</pubDate></item>
      <item><title>Unrelated</title><link>javascript:alert(1)</link><description>Bad link</description></item>
      <item><title>Renovación de cocina</title><link>/obras/2</link><description>Reparación</description></item>
    </channel></rss>`
	got, err := ExtractFeedCandidates([]byte(rss), "https://example.com/feed")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 safe items, got %d: %+v", len(got), got)
	}
	if got[0].PublishedAt == nil || !RelevantRenovationCandidate(got[0]) {
		t.Fatalf("missing date or relevant keyword: %+v", got[0])
	}
	if got[1].URL != "https://example.com/obras/2" {
		t.Fatalf("relative URL not resolved: %s", got[1].URL)
	}
	atom := `<feed xmlns="http://www.w3.org/2005/Atom"><entry><title>Obras de reforma</title>
      <link rel="self" href="https://example.com/feed"/><link rel="alternate" href="https://example.com/trabajos/9"/>
      <updated>2026-11-20T09:00:00Z</updated><summary>Construcción</summary></entry></feed>`
	got, err = ExtractFeedCandidates([]byte(atom), "https://example.com/feed")
	if err != nil || len(got) != 1 || got[0].URL != "https://example.com/trabajos/9" {
		t.Fatalf("invalid Atom entry: %+v, %v", got, err)
	}
}

func TestExtractFeedCandidatesRejectsInvalidInputs(t *testing.T) {
	for _, raw := range []string{"", "<html>bad</html>", "<rss", strings.Repeat("x", 2*1024*1024+1)} {
		if _, err := ExtractFeedCandidates([]byte(raw), "https://example.com/feed"); err == nil {
			t.Fatalf("must reject invalid feed of length %d", len(raw))
		}
	}
}
