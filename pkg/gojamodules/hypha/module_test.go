package hyphamod_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dop251/goja"
	gggengine "github.com/go-go-golems/go-go-goja/pkg/engine"
	_ "github.com/go-go-golems/hypha-cli/pkg/gojamodules/hypha" // register the module
	"github.com/stretchr/testify/require"
)

// newRuntimeWithBackend starts an httptest server that proxies to a handler
// and returns a go-go-goja runtime whose "hypha" module is configured to use
// it. We inject the base URL via SetModuleConfig.
func newRuntimeWithBackend(t *testing.T, fn func(w http.ResponseWriter, r *http.Request)) (*gggengine.Runtime, string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", fn)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	factory, err := gggengine.NewRuntimeFactoryBuilder().Build()
	require.NoError(t, err)
	rt, err := factory.NewRuntime(
		gggengine.WithStartupContext(context.Background()),
		gggengine.WithLifetimeContext(context.Background()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = rt.Close(context.Background()) })
	return rt, srv.URL
}

func runJS(t *testing.T, rt *gggengine.Runtime, code string) string {
	t.Helper()
	val, err := rt.Owner.Call(context.Background(), "test.run", func(_ context.Context, vm *goja.Runtime) (any, error) {
		v, err := vm.RunString(code)
		if err != nil {
			return nil, err
		}
		return v.Export(), nil
	})
	require.NoError(t, err)
	s, ok := val.(string)
	require.Truef(t, ok, "expected string, got %T", val)
	return s
}

func runJSErr(t *testing.T, rt *gggengine.Runtime, code string) error {
	t.Helper()
	_, err := rt.Owner.Call(context.Background(), "test.run", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, err := vm.RunString(code)
		return nil, err
	})
	return err
}

// setConfig injects the base URL + PAT into the module before require().
func setConfig(t *testing.T, rt *gggengine.Runtime, baseURL, pat string) {
	t.Helper()
	_, err := rt.Owner.Call(context.Background(), "test.setconfig", func(_ context.Context, vm *goja.Runtime) (any, error) {
		// Use the exported SetModuleConfig via a tiny JS shim is not possible
		// (it's a Go symbol); instead set env vars which the Loader reads.
		return nil, nil
	})
	require.NoError(t, err)
	t.Setenv("HYPHA_BASE_URL", baseURL)
	t.Setenv("HYPHA_PAT", pat)
}

func TestRequireHypha(t *testing.T) {
	rt, _ := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {})
	got := runJS(t, rt, `typeof require("hypha")`)
	require.Equal(t, "object", got)
}

func TestEventBuilderPost(t *testing.T) {
	rt, base := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/events" {
			t.Fatalf("got %s %s", r.Method, r.URL.Path)
		}
		io.WriteString(w, `{"id":"01ABC","ts":1,"actor":"01","verb":"post","kind":"update","target":null,"ref":null,"audience":"circle","topics":["meta"],"body":"hi","value_amount":null,"value_unit":null,"redacted":0}`)
	})
	setConfig(t, rt, base, "hh_pat_test")
	got := runJS(t, rt, `
		const h = require("hypha");
		const ev = h.event("hi").Kind("update").Topic("meta").Post();
		JSON.stringify({ id: ev.ID, verb: ev.Verb, kind: ev.Kind, topics: ev.Topics });
	`)
	require.JSONEq(t, `{"id":"01ABC","verb":"post","kind":"update","topics":["meta"]}`, got)
}

func TestEventBuilderHoursDebtAudienceWarning(t *testing.T) {
	rt, _ := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not call the server when validation fails")
	})
	setConfig(t, rt, "http://example.invalid", "hh_pat_test")
	// Non-circle audience + hours + a target -> the audience-debt warning fires.
	err := runJSErr(t, rt, `
		require("hypha").event("gave 2h").Kind("gig").To("@alice").Audience("01OTHER").Value(2, "hours").Post();
	`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "value: unit \"hours\" creates time debt")
	require.Contains(t, err.Error(), "audience should be \"circle\"")
}

func TestEventBuilderValidationAggregatesErrors(t *testing.T) {
	rt, _ := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not call the server when validation fails")
	})
	setConfig(t, rt, "http://example.invalid", "hh_pat_test")
	err := runJSErr(t, rt, `
		require("hypha").event("gave alice 2h").Kind("gig").Value(2, "hours").Post();
	`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "value: a valued event requires .to(<member>)")
}

func TestEventBuilderValidateReturnsGoResult(t *testing.T) {
	rt, _ := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {})
	setConfig(t, rt, "http://example.invalid", "hh_pat_test")
	got := runJS(t, rt, `
		const v = require("hypha").event("x").Value(2, "hours").Validate();
		JSON.stringify({ valid: v.Valid, n: v.Errors.length });
	`)
	require.JSONEq(t, `{"valid":false,"n":1}`, got) // missing target only (audience defaults to circle)
}

func TestFeedBuilderList(t *testing.T) {
	rt, base := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("topic") != "meta" {
			t.Fatalf("topic query = %q", r.URL.Query().Get("topic"))
		}
		io.WriteString(w, `{"events":[{"id":"01A","ts":1,"actor":"x","verb":"post","audience":"circle","topics":[],"redacted":0}]}`)
	})
	setConfig(t, rt, base, "hh_pat_test")
	got := runJS(t, rt, `
		const evs = require("hypha").feed().Topic("meta").List();
		JSON.stringify({ n: evs.length, id: evs[0].ID });
	`)
	require.JSONEq(t, `{"n":1,"id":"01A"}`, got)
}

func TestViewsBalance(t *testing.T) {
	rt, base := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"members":[{"id":"01","handle":"alice","name":"Alice","received":0,"given":0,"balance":0}]}`)
	})
	setConfig(t, rt, base, "hh_pat_test")
	got := runJS(t, rt, `
		const rows = require("hypha").views.balance();
		JSON.stringify({ n: rows.length, handle: rows[0].Handle });
	`)
	require.JSONEq(t, `{"n":1,"handle":"alice"}`, got)
}

func TestPatsBuilderMint(t *testing.T) {
	rt, base := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/pats" {
			t.Fatalf("got %s %s", r.Method, r.URL.Path)
		}
		io.WriteString(w, `{"token":"hh_pat_new","id":"01P"}`)
	})
	setConfig(t, rt, base, "hh_pat_test")
	got := runJS(t, rt, `
		const res = require("hypha").pats().Scopes("read").Label("agent").Mint();
		JSON.stringify({ token: res.Token, id: res.ID });
	`)
	require.JSONEq(t, `{"token":"hh_pat_new","id":"01P"}`, got)
}

func TestPatsBuilderMintRequiresScopes(t *testing.T) {
	rt, _ := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not call server")
	})
	setConfig(t, rt, "http://example.invalid", "hh_pat_test")
	err := runJSErr(t, rt, `require("hypha").pats().Mint();`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least one scope is required")
}

func TestCheckpointsBuilder(t *testing.T) {
	rt, base := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"checkpoints":[{"seq":1,"up_to_id":"01A","event_count":3,"hash":"abc","prev_hash":null,"created_at":1}]}`)
	})
	setConfig(t, rt, base, "hh_pat_test")
	got := runJS(t, rt, `
		const cps = require("hypha").checkpoints().Limit(5).List();
		JSON.stringify({ n: cps.length, hash: cps[0].Hash });
	`)
	require.JSONEq(t, `{"n":1,"hash":"abc"}`, got)
}

func TestUnknownMethodIsTypeError(t *testing.T) {
	rt, _ := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {})
	setConfig(t, rt, "http://example.invalid", "hh_pat_test")
	err := runJSErr(t, rt, `require("hypha").event("x").toppic("meta");`)
	require.Error(t, err)
	// goja reports the unknown method as a TypeError.
	require.Contains(t, err.Error(), "has no member 'toppic'")
}

func TestToSpecSerialization(t *testing.T) {
	rt, _ := newRuntimeWithBackend(t, func(w http.ResponseWriter, r *http.Request) {})
	setConfig(t, rt, "http://example.invalid", "hh_pat_test")
	got := runJS(t, rt, `
		const spec = require("hypha").event("hi").Kind("update").Topic("meta").ToSpec();
		JSON.stringify(spec);
	`)
	// Just assert it's valid JSON with the expected keys.
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(got), &m))
	require.Equal(t, "hi", m["body"])
	require.Equal(t, "update", m["kind"])
}
