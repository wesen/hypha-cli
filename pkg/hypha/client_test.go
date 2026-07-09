package hypha_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-go-golems/hypha-cli/pkg/hypha"
)

// newServer returns a test server and a client pointed at it. The handler
// asserts the bearer token and records the last request for inspection.
func newServer(t *testing.T, fn func(w http.ResponseWriter, r *http.Request)) (*hypha.Client, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", fn)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := hypha.NewClient(srv.URL, "hh_pat_test")
	return c, srv
}

func assertAuth(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.Header.Get("Authorization"); got != "Bearer hh_pat_test" {
		t.Fatalf("auth header = %q, want Bearer hh_pat_test", got)
	}
}

func TestPost(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/events" {
			t.Fatalf("got %s %s, want POST /api/v1/events", r.Method, r.URL.Path)
		}
		var o hypha.PostOptions
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			t.Fatal(err)
		}
		if o.Body != "hi" || o.Kind != "update" || len(o.Topics) != 1 || o.Topics[0] != "meta" {
			t.Fatalf("unexpected post opts: %+v", o)
		}
		b := `{"id":"01ABC","ts":1700000000,"actor":"01ME","verb":"post","kind":"update","target":null,"ref":null,"audience":"circle","topics":["meta"],"body":"hi","value_amount":null,"value_unit":null,"redacted":0}`
		io.WriteString(w, b)
	})
	ev, err := c.Post(context.Background(), hypha.PostOptions{Body: "hi", Kind: "update", Topics: []string{"meta"}, Audience: "circle"})
	if err != nil {
		t.Fatal(err)
	}
	if ev.ID != "01ABC" || ev.Verb != "post" || ev.Audience != "circle" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	if ev.Kind == nil || *ev.Kind != "update" {
		t.Fatalf("kind = %v", ev.Kind)
	}
}

func TestFeedEnvelope(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.URL.Path != "/api/v1/events" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("topic") != "meta" {
			t.Fatalf("topic query = %q", r.URL.Query().Get("topic"))
		}
		io.WriteString(w, `{"events":[{"id":"01A","ts":1,"actor":"x","verb":"post","kind":null,"target":null,"ref":null,"audience":"circle","topics":[],"body":null,"value_amount":null,"value_unit":null,"redacted":0}]}`)
	})
	evs, err := c.Feed(context.Background(), hypha.FeedOptions{Topics: []string{"meta"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 || evs[0].ID != "01A" {
		t.Fatalf("feed = %+v", evs)
	}
}

func TestMembersEnvelope(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.URL.Path != "/api/v1/members" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{"members":[{"id":"01","handle":"alice","name":"Alice","status":"active"}]}`)
	})
	ms, err := c.Members(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 1 || ms[0].Handle != "alice" {
		t.Fatalf("members = %+v", ms)
	}
}

func TestWhoisWithTrustProfile(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.URL.Path != "/api/v1/members/@alice" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{"id":"01","handle":"alice","name":"Alice","status":"active","trust":{"balance":2,"hours_given":1,"hours_received":3,"invites_given":0,"connections":1},"profile":{"bio_long":null,"links":[],"skills":["go"],"open_to_work":null,"location_note":"","last_active":1700000000}}`)
	})
	m, err := c.Whois(context.Background(), "@alice")
	if err != nil {
		t.Fatal(err)
	}
	if m.Trust == nil || m.Trust.HoursReceived != 3 {
		t.Fatalf("trust = %+v", m.Trust)
	}
	if m.Profile == nil || len(m.Profile.Skills) != 1 || m.Profile.Skills[0] != "go" {
		t.Fatalf("profile = %+v", m.Profile)
	}
}

func TestMintAndListPATs(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		switch r.URL.Path {
		case "/api/v1/pats":
			if r.Method == http.MethodPost {
				io.WriteString(w, `{"token":"hh_pat_deadbeef","id":"01P"}`)
				return
			}
			io.WriteString(w, `{"pats":[{"id":"01P","scopes":["read"],"label":"agent","created_at":1,"revoked_at":null}]}`)
			return
		}
		t.Fatalf("unexpected path %s", r.URL.Path)
	})
	res, err := c.MintPAT(context.Background(), hypha.MintPATOptions{Scopes: []string{"read"}, Label: "agent"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Token != "hh_pat_deadbeef" || res.ID != "01P" {
		t.Fatalf("mint = %+v", res)
	}
	pats, err := c.ListPATs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(pats) != 1 || pats[0].Label != "agent" {
		t.Fatalf("pats = %+v", pats)
	}
}

func TestCheckpointsEnvelope(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		io.WriteString(w, `{"checkpoints":[{"seq":1,"up_to_id":"01A","event_count":3,"hash":"abc","prev_hash":null,"created_at":1}]}`)
	})
	cps, err := c.Checkpoints(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(cps) != 1 || cps[0].Hash != "abc" || cps[0].EventCount != 3 {
		t.Fatalf("checkpoints = %+v", cps)
	}
}

func TestErrorShape(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"error":"not found"}`)
	})
	_, err := c.Whois(context.Background(), "@nope")
	if err == nil {
		t.Fatal("want error")
	}
	he, ok := err.(*hypha.Error)
	if !ok {
		t.Fatalf("err type = %T", err)
	}
	if he.Status != http.StatusNotFound || he.Message != "not found" {
		t.Fatalf("err = %+v", he)
	}
}

func TestExportPagination(t *testing.T) {
	calls := 0
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		calls++
		if calls == 1 {
			io.WriteString(w, `{"member":{"id":"01","handle":"alice","name":"Alice","status":"active"},"balance":{"received":0,"given":0,"balance":0},"edges":[],"events":[{"id":"01A","ts":1,"actor":"01","verb":"post","audience":"circle","topics":[],"redacted":0}],"next_cursor":"02A"}`)
			return
		}
		if r.URL.Query().Get("cursor") != "02A" {
			t.Fatalf("cursor = %q", r.URL.Query().Get("cursor"))
		}
		io.WriteString(w, `{"events":[{"id":"02A","ts":2,"actor":"01","verb":"post","audience":"circle","topics":[],"redacted":0}],"next_cursor":null}`)
	})
	ex, err := c.ExportAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ex.Events) != 2 || ex.Events[0].ID != "01A" || ex.Events[1].ID != "02A" {
		t.Fatalf("events = %+v", ex.Events)
	}
	if ex.Member == nil || ex.Member.Handle != "alice" {
		t.Fatalf("member = %+v", ex.Member)
	}
}

func TestVerifyChain(t *testing.T) {
	// Two events; compute the expected chain in the test using the same
	// canonical format the client uses, then assert Verify says OK.
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		// Return two events, no next cursor.
		io.WriteString(w, `{"events":[{"id":"01A","ts":1000,"actor":"01","verb":"post","kind":null,"target":null,"ref":null,"audience":"circle","topics":[],"body":null,"value_amount":null,"value_unit":null,"redacted":0},{"id":"01B","ts":2000,"actor":"01","verb":"post","kind":null,"target":null,"ref":null,"audience":"circle","topics":[],"body":null,"value_amount":null,"value_unit":null,"redacted":0}],"next_cursor":null}`)
	})
	// Compute expected running hash.
	running := []byte("")
	for _, s := range []string{"01A|1000|01|post|||", "01B|2000|01|post|||"} {
		h := sha256.New()
		h.Write(running)
		h.Write([]byte(s))
		running = h.Sum(nil)
	}
	res, err := c.Verify(context.Background(), "01B", hex.EncodeToString(running))
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("verify not OK; computed=%s", res.ComputedHash)
	}
	if res.EventCount != 2 {
		t.Fatalf("count = %d", res.EventCount)
	}
}
