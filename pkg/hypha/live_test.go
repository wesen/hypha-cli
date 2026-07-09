package hypha_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-go-golems/hypha-cli/pkg/hypha"
)

// Live integration tests against $HYPHA_BASE_URL with $HYPHA_PAT. Skipped
// unless both are set. These are read-only (plus one post+redact round trip)
// to avoid polluting the circle.

func liveClient(t *testing.T) *hypha.Client {
	t.Helper()
	base := os.Getenv("HYPHA_BASE_URL")
	pat := os.Getenv("HYPHA_PAT")
	if base == "" || pat == "" {
		t.Skip("HYPHA_BASE_URL/HYPHA_PAT not set; skipping live test")
	}
	return hypha.NewClient(base, pat)
}

func TestLiveHealth(t *testing.T) {
	c := liveClient(t)
	s, err := c.Health(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if s != "ok" {
		t.Fatalf("health = %q", s)
	}
}

func TestLiveMembersAndWhois(t *testing.T) {
	c := liveClient(t)
	ms, err := c.Members(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) == 0 {
		t.Fatal("no members")
	}
	// Whois the first member by handle.
	h := "@" + ms[0].Handle
	m, err := c.Whois(context.Background(), h)
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != ms[0].ID {
		t.Fatalf("whois id %q != members id %q", m.ID, ms[0].ID)
	}
}

func TestLiveFeed(t *testing.T) {
	c := liveClient(t)
	evs, err := c.Feed(context.Background(), hypha.FeedOptions{Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	// Feed may be empty; just assert no error and decode worked.
	_ = evs
}

func TestLiveBalanceAndTopics(t *testing.T) {
	c := liveClient(t)
	if _, err := c.Balance(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Topics(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestLivePatsList(t *testing.T) {
	c := liveClient(t)
	if _, err := c.ListPATs(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestLivePostAndRedact(t *testing.T) {
	c := liveClient(t)
	ev, err := c.Post(context.Background(), hypha.PostOptions{
		Body:     "hypha-cli live test (will redact)",
		Kind:     "acme.probe",
		Topics:   []string{"meta"},
		Audience: "circle",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Redact(context.Background(), ev.ID); err != nil {
		t.Fatal(err)
	}
}

func TestLiveCheckpointsAndVerify(t *testing.T) {
	c := liveClient(t)
	cps, err := c.Checkpoints(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(cps) == 0 {
		t.Skip("no checkpoints yet")
	}
	// Verify recomputes the chain over the CALLER's exported events (export is
	// scoped to the caller). A global checkpoint covers events authored by
	// everyone, so OK may be false for a non-author. We only assert that Verify
	// runs and produces a deterministic computed hash.
	cp := cps[0]
	res, err := c.Verify(context.Background(), cp.UpToID, cp.Hash)
	if err != nil {
		t.Fatal(err)
	}
	if res.ComputedHash == "" && res.EventCount > 0 {
		t.Fatal("empty computed hash with events")
	}
	t.Logf("verify ok=%v computed=%s events=%d", res.OK, res.ComputedHash, res.EventCount)
}
