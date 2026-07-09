package hyphamod

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dop251/goja"
	"github.com/go-go-golems/hypha-cli/pkg/hypha"
)

// FeedQuery is the fluent builder for the event feed.
type FeedQuery struct {
	rt    *runtime
	opts  hypha.FeedOptions
}

func (r *runtime) feed() *FeedQuery { return &FeedQuery{rt: r} }

func (q *FeedQuery) Kind(k string) *FeedQuery      { q.opts.Kind = k; return q }
func (q *FeedQuery) Topic(t ...string) *FeedQuery  { q.opts.Topics = append(q.opts.Topics, t...); return q }
func (q *FeedQuery) Actor(a string) *FeedQuery     { q.opts.Actor = a; return q }
func (q *FeedQuery) Ref(ref string) *FeedQuery    { q.opts.Ref = ref; return q }
func (q *FeedQuery) Since(s string) *FeedQuery    { q.opts.Since = s; return q }
func (q *FeedQuery) Limit(n int) *FeedQuery       { q.opts.Limit = n; return q }

func (q *FeedQuery) List() ([]hypha.Event, error) {
	return q.rt.client.Feed(context.Background(), q.opts)
}

// SearchQuery is the fluent builder for full-text search.
type SearchQuery struct {
	rt    *runtime
	q     string
	limit int
}

func (r *runtime) search(q string) *SearchQuery { return &SearchQuery{rt: r, q: q} }

func (s *SearchQuery) Limit(n int) *SearchQuery { s.limit = n; return s }

func (s *SearchQuery) List() ([]hypha.Event, error) {
	return s.rt.client.Search(context.Background(), s.q, s.limit)
}

// --- Identity (terminals return Go-backed results) ---

func (r *runtime) whois(id string) (*hypha.Member, error) {
	return r.client.Whois(context.Background(), id)
}

func (r *runtime) whoami() (*hypha.Member, error) {
	return r.client.Whoami(context.Background(), r.handle)
}

func (r *runtime) members() ([]hypha.Member, error) {
	return r.client.Members(context.Background())
}

func (r *runtime) kudos(target, body string) (*hypha.Event, error) {
	return r.client.Post(context.Background(), hypha.PostOptions{
		Kind: "kudos", Target: target, Body: body,
		Value: &hypha.Value{Amount: 1, Unit: "kudos"},
	})
}

// --- Views (terminals return Go-backed results) ---

func (r *runtime) balance() ([]hypha.BalanceRow, error) {
	return r.client.Balance(context.Background())
}
func (r *runtime) trust() (*hypha.TrustResult, error) {
	return r.client.Trust(context.Background())
}
func (r *runtime) topics() ([]hypha.TopicRow, error) {
	return r.client.Topics(context.Background())
}
func (r *runtime) whoCanHelp(topic string) ([]hypha.WhoCanHelpRow, error) {
	return r.client.WhoCanHelp(context.Background(), topic)
}
func (r *runtime) gratitude(member string) (*hypha.GratitudeResult, error) {
	return r.client.Gratitude(context.Background(), member)
}

// --- PATs builder ---

type PATBuilder struct {
	rt     *runtime
	scopes []string
	label  string
}

func (r *runtime) pats() *PATBuilder { return &PATBuilder{rt: r} }

func (b *PATBuilder) Scopes(scopes ...string) *PATBuilder { b.scopes = scopes; return b }
func (b *PATBuilder) Label(l string) *PATBuilder          { b.label = l; return b }

func (b *PATBuilder) Mint() (*hypha.MintPATResult, error) {
	if len(b.scopes) == 0 {
		return nil, fmt.Errorf("hypha.pats.mint: at least one scope is required")
	}
	return b.rt.client.MintPAT(context.Background(), hypha.MintPATOptions{Scopes: b.scopes, Label: b.label})
}

// --- Webhooks ---

type WebhookBuilder struct {
	rt     *runtime
	url    string
	secret string
}

func (r *runtime) webhook(url string) *WebhookBuilder { return &WebhookBuilder{rt: r, url: url} }

func (b *WebhookBuilder) Secret(s string) *WebhookBuilder { b.secret = s; return b }

func (b *WebhookBuilder) Subscribe() (*hypha.Webhook, error) {
	return b.rt.client.Subscribe(context.Background(), hypha.SubscribeOptions{URL: b.url, Secret: b.secret})
}

func (r *runtime) webhooksList() ([]hypha.Webhook, error) {
	return r.client.Webhooks(context.Background())
}

func (r *runtime) unsubscribe(id string) error {
	return r.client.Unsubscribe(context.Background(), id)
}

// --- Export / Checkpoints / Verify ---

type ExportQuery struct{ rt *runtime }

func (r *runtime) exportQuery() *ExportQuery { return &ExportQuery{rt: r} }

func (q *ExportQuery) All() (*hypha.Export, error) {
	return q.rt.client.ExportAll(context.Background())
}

type CheckpointsQuery struct {
	rt    *runtime
	limit int
}

func (r *runtime) checkpointsQuery() *CheckpointsQuery { return &CheckpointsQuery{rt: r} }

func (q *CheckpointsQuery) Limit(n int) *CheckpointsQuery { q.limit = n; return q }

func (q *CheckpointsQuery) List() ([]hypha.Checkpoint, error) {
	return q.rt.client.Checkpoints(context.Background(), q.limit)
}

type VerifyQuery struct {
	rt   *runtime
	upTo string
	hash string
}

func (r *runtime) verifyQuery() *VerifyQuery { return &VerifyQuery{rt: r} }

func (q *VerifyQuery) UpTo(id string) *VerifyQuery { q.upTo = id; return q }
func (q *VerifyQuery) Hash(h string) *VerifyQuery   { q.hash = h; return q }

func (q *VerifyQuery) Run() (*hypha.VerifyResult, error) {
	if q.upTo == "" || q.hash == "" {
		return nil, fmt.Errorf("hypha.verify: up-to and hash are required")
	}
	return q.rt.client.Verify(context.Background(), q.upTo, q.hash)
}

// objectToStringSlice reads a JS array into []string.
func objectToStringSlice(vm *goja.Runtime, v goja.Value) []string {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return nil
	}
	exp := v.Export()
	arr, ok := exp.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		out = append(out, fmt.Sprintf("%v", e))
	}
	return out
}

var _ = strconv.Itoa

// --- Asks (ISO/gig) via MCP /mcp tools ---
//
// The ask domain (ISO board: post_iso, reply_ask, close_ask, list_asks,
// get_ask) lives behind the MCP endpoint, not the REST /api/v1/events log.
// These methods call client.MCP() and return Go-backed results.

func (r *runtime) postISO(body string, topics []string) (*hypha.Ask, error) {
	return r.client.MCP().PostISO(context.Background(), hypha.PostISOOptions{Body: body, Topics: topics})
}

func (r *runtime) replyAsk(id, body string) (*hypha.AskReply, error) {
	return r.client.MCP().ReplyAsk(context.Background(), id, body)
}

func (r *runtime) closeAsk(id, helped, note string) error {
	return r.client.MCP().CloseAsk(context.Background(), hypha.CloseAskOptions{ID: id, Helped: helped, Note: note})
}

func (r *runtime) listAsks(flavor, status string) (*hypha.AskList, error) {
	return r.client.MCP().ListAsks(context.Background(), flavor, status)
}

func (r *runtime) getAsk(id string) (*hypha.AskThread, error) {
	return r.client.MCP().GetAsk(context.Background(), id)
}
