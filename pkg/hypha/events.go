package hypha

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Value is the {amount, unit} pair attached to a valued event. Units are
// namespaced: "hours" (time ledger, debt-creating), "kudos" (abundant), or
// dot-prefixed third-party units. Amounts are numbers (hours may be
// fractional; tips are integers in minor units).
type Value struct {
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}

// Event mirrors the kernel store schema (see docs/portability "Event fields").
// kind/target/ref/value_* are null when absent, hence pointers.
type Event struct {
	ID          string   `json:"id"`
	Ts          int64    `json:"ts"`
	Actor       string   `json:"actor"`
	Verb        string   `json:"verb"`
	Kind        *string  `json:"kind"`
	Target      *string  `json:"target"`
	Ref         *string  `json:"ref"`
	Audience    string   `json:"audience"`
	Topics      []string `json:"topics"`
	Body        *string  `json:"body"`
	ValueAmount *float64 `json:"value_amount"`
	ValueUnit   *string  `json:"value_unit"`
	Redacted    int      `json:"redacted"`
}

// ValueAmount returns the value amount or 0 when absent.
func (e *Event) ValueAmountOrZero() float64 {
	if e == nil || e.ValueAmount == nil {
		return 0
	}
	return *e.ValueAmount
}

// ValueUnitOrEmpty returns the value unit or "" when absent.
func (e *Event) ValueUnitOrEmpty() string {
	if e == nil || e.ValueUnit == nil {
		return ""
	}
	return *e.ValueUnit
}

// PostOptions is the body of POST /api/v1/events (the universal append).
// All optional fields are omitted from the JSON when zero.
type PostOptions struct {
	Body     string   `json:"body,omitempty"`
	Kind     string   `json:"kind,omitempty"`
	Topics   []string `json:"topics,omitempty"`
	Target   string   `json:"target,omitempty"`   // member id or @handle
	Ref      string   `json:"ref,omitempty"`      // referenced event id
	Audience string   `json:"audience,omitempty"` // "circle" (default) or a member id
	Value    *Value   `json:"value,omitempty"`
	IdemKey  string   `json:"idem_key,omitempty"`
}

// Post appends an event and returns the created Event.
func (c *Client) Post(ctx context.Context, o PostOptions) (*Event, error) {
	var ev Event
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/events", o, &ev); err != nil {
		return nil, err
	}
	return &ev, nil
}

// FeedOptions filters the event feed. These map to query params.
type FeedOptions struct {
	Kind   string
	Topics []string
	Actor  string
	Ref    string
	Since  string
	Limit  int
}

// Feed returns the event feed (GET /api/v1/events). Response envelope:
// {"events":[...]}.
func (c *Client) Feed(ctx context.Context, o FeedOptions) ([]Event, error) {
	params := url.Values{}
	if o.Kind != "" {
		params.Set("kind", o.Kind)
	}
	for _, t := range o.Topics {
		params.Add("topic", t)
	}
	if o.Actor != "" {
		params.Set("actor", o.Actor)
	}
	if o.Ref != "" {
		params.Set("ref", o.Ref)
	}
	if o.Since != "" {
		params.Set("since", o.Since)
	}
	if o.Limit > 0 {
		params.Set("limit", strconv.Itoa(o.Limit))
	}
	var wrap struct {
		Events []Event `json:"events"`
	}
	if err := c.doGet(ctx, "/api/v1/events", params, &wrap); err != nil {
		return nil, err
	}
	return wrap.Events, nil
}

// SearchHit is a result row from GET /api/v1/search. It is an Event.
type SearchHit = Event

// Search runs full-text search over non-redacted, viewer-visible bodies.
// Response envelope: {"events":[...]}.
func (c *Client) Search(ctx context.Context, q string, limit int) ([]SearchHit, error) {
	params := url.Values{}
	params.Set("q", q)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	var wrap struct {
		Events []Event `json:"events"`
	}
	if err := c.doGet(ctx, "/api/v1/search", params, &wrap); err != nil {
		return nil, err
	}
	return wrap.Events, nil
}

// Redact blanks an event's body/topics (POST /api/v1/events/:id/redact).
// The kernel keeps the immutable fact (actor, verb, value, ts).
func (c *Client) Redact(ctx context.Context, eventID string) error {
	var ok struct {
		OK bool `json:"ok"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/events/"+url.PathEscape(eventID)+"/redact", nil, &ok); err != nil {
		return err
	}
	return nil
}
