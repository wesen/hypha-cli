package hypha

import (
	"context"
	"net/url"
)

// ExportPage is one page of GET /api/v1/export. The first page includes
// Member, Balance, Edges, and Profile; subsequent pages omit them and only
// continue Events from the cursor. Events are oldest-first (ULID order).
type ExportPage struct {
	Member     *Member        `json:"member,omitempty"`
	Balance    *ExportBalance `json:"balance,omitempty"`
	Edges      []TrustEdge    `json:"edges,omitempty"`
	Profile    *Profile       `json:"profile,omitempty"`
	Events     []Event        `json:"events"`
	NextCursor *string        `json:"next_cursor"` // nil on the last page
}

// ExportBalance is the balance sub-object on the first export page.
type ExportBalance struct {
	Received float64 `json:"received"`
	Given    float64 `json:"given"`
	Balance  float64 `json:"balance"`
}

// Export fetches one page of the caller's event history. Pass cursor="" (or
// omit) for the first page; follow NextCursor until it is nil.
func (c *Client) Export(ctx context.Context, cursor string) (*ExportPage, error) {
	params := url.Values{}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	var page ExportPage
	if err := c.doGet(ctx, "/api/v1/export", params, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Export is the fully-aggregated export (all pages merged).
type Export struct {
	Member  *Member
	Balance *ExportBalance
	Edges   []TrustEdge
	Profile *Profile
	Events  []Event
}

// ExportAll follows every export page until NextCursor is nil and returns
// one combined Export (member/balance/edges/profile from the first page,
// all events concatenated).
func (c *Client) ExportAll(ctx context.Context) (*Export, error) {
	out := &Export{}
	cursor := ""
	first := true
	for {
		page, err := c.Export(ctx, cursor)
		if err != nil {
			return nil, err
		}
		if first {
			out.Member, out.Balance, out.Edges, out.Profile = page.Member, page.Balance, page.Edges, page.Profile
			first = false
		}
		out.Events = append(out.Events, page.Events...)
		if page.NextCursor == nil {
			return out, nil
		}
		cursor = *page.NextCursor
	}
}
