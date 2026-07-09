package hypha

import (
	"context"
	"net/url"
)

// BalanceRow is one member's time balance (hours received/given/net).
type BalanceRow struct {
	ID       string  `json:"id"`
	Handle   string  `json:"handle"`
	Name     string  `json:"name"`
	Received float64 `json:"received"`
	Given    float64 `json:"given"`
	Balance  float64 `json:"balance"`
}

// TrustEdge is one trust-graph edge for the caller.
type TrustEdge struct {
	Peer   string `json:"peer"`
	Type   string `json:"type"` // "helped" | "vouched" | "introduced"
	Weight int    `json:"weight"`
	LastTs int64  `json:"last_ts"`
}

// TrustResult is the trust view: the caller's member id plus edges.
type TrustResult struct {
	Member string      `json:"member"`
	Edges  []TrustEdge `json:"edges"`
}

// TopicRow is one active topic in the circle.
type TopicRow struct {
	Topic        string `json:"topic"`
	Events       int    `json:"events"`
	Actors       int    `json:"actors"`
	LastEventID  string `json:"last_event_id"`
}

// WhoCanHelpRow is one member ranked by helpfulness on a topic.
type WhoCanHelpRow struct {
	Member     string  `json:"member"`
	Posts      int     `json:"posts"`
	HoursGiven float64 `json:"hours_given"`
	Balance    float64 `json:"balance"`
	Score      float64 `json:"score"`
}

// GratitudeRow is one member's gratitude score.
type GratitudeRow struct {
	Member        string `json:"member"`
	Handle        string `json:"handle,omitempty"`
	KudosReceived int    `json:"kudos_received"`
	Gratitude     int    `json:"gratitude"`
}

// GratitudeResult wraps the gratitude view (all members or one).
type GratitudeResult struct {
	Members []GratitudeRow `json:"members"`
}

// Balance returns every active member's time balance.
func (c *Client) Balance(ctx context.Context) ([]BalanceRow, error) {
	var wrap struct {
		Members []BalanceRow `json:"members"`
	}
	if err := c.doGet(ctx, "/api/v1/views/balance", nil, &wrap); err != nil {
		return nil, err
	}
	return wrap.Members, nil
}

// Trust returns the caller's trust edges.
func (c *Client) Trust(ctx context.Context) (*TrustResult, error) {
	var res TrustResult
	if err := c.doGet(ctx, "/api/v1/views/trust", nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Topics returns the top 100 active circle topics.
func (c *Client) Topics(ctx context.Context) ([]TopicRow, error) {
	var wrap struct {
		Topics []TopicRow `json:"topics"`
	}
	if err := c.doGet(ctx, "/api/v1/views/topics", nil, &wrap); err != nil {
		return nil, err
	}
	return wrap.Topics, nil
}

// WhoCanHelp returns up to 20 members ranked by helpfulness on a topic.
func (c *Client) WhoCanHelp(ctx context.Context, topic string) ([]WhoCanHelpRow, error) {
	params := url.Values{}
	params.Set("topic", topic)
	var wrap struct {
		Members []WhoCanHelpRow `json:"members"`
	}
	if err := c.doGet(ctx, "/api/v1/views/who-can-help", params, &wrap); err != nil {
		return nil, err
	}
	return wrap.Members, nil
}

// Gratitude returns kudos-based gratitude scores. If member is empty, all
// members are returned; otherwise the single member's score.
func (c *Client) Gratitude(ctx context.Context, member string) (*GratitudeResult, error) {
	params := url.Values{}
	if member != "" {
		params.Set("member", member)
	}
	var res GratitudeResult
	if err := c.doGet(ctx, "/api/v1/views/gratitude", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
