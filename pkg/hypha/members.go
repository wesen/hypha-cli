package hypha

import (
	"context"
	"net/url"
)

// Member is a member identity. The whois endpoint embeds Trust and Profile;
// the members-list endpoint returns the base fields only.
type Member struct {
	ID        string `json:"id"`
	Handle    string `json:"handle"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	Timezone  string `json:"timezone"`
	Status    string `json:"status"`
	InvitedBy string `json:"invited_by"`
	CreatedAt int64  `json:"created_at"`

	// Trust and Profile are populated by whois only (nil on list members).
	Trust   *TrustSummary `json:"trust,omitempty"`
	Profile *Profile     `json:"profile,omitempty"`
}

// TrustSummary is the trust sub-object on a whois response.
type TrustSummary struct {
	Balance       int `json:"balance"`
	HoursGiven     int `json:"hours_given"`
	HoursReceived int `json:"hours_received"`
	InvitesGiven  int `json:"invites_given"`
	Connections   int `json:"connections"`
}

// Profile is the profile sub-object on a whois response.
type Profile struct {
	BioLong      string   `json:"bio_long"`
	Links        []string `json:"links"`
	Skills       []string `json:"skills"`
	// OpenToWork is a string (e.g. "pt", "ft") or null — the live API returns
// a string, not a bool as the docs implied.
OpenToWork   *string  `json:"open_to_work"`
	LocationNote string   `json:"location_note"`
	LastActive   int64    `json:"last_active"`
}

// Members lists all members (GET /api/v1/members). Response envelope:
// {"members":[...]}.
func (c *Client) Members(ctx context.Context) ([]Member, error) {
	var wrap struct {
		Members []Member `json:"members"`
	}
	if err := c.doGet(ctx, "/api/v1/members", nil, &wrap); err != nil {
		return nil, err
	}
	return wrap.Members, nil
}

// Whois returns a single member by id or @handle (GET /api/v1/members/:idOrHandle).
func (c *Client) Whois(ctx context.Context, idOrHandle string) (*Member, error) {
	var m Member
	if err := c.doGet(ctx, "/api/v1/members/"+url.PathEscape(idOrHandle), nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Whoami resolves the caller's own member. The kernel has no @me endpoint,
// so this resolves via a configured handle (preferred) or, if handle is empty,
// by listing members and returning the first whose status is active and
// whose handle matches the config (callers should pass a handle).
func (c *Client) Whoami(ctx context.Context, handle string) (*Member, error) {
	if handle != "" {
		return c.Whois(ctx, handle)
	}
	// Fallback: cannot resolve without a handle; return a clear error.
	return nil, &Error{Status: 0, Message: "whoami requires a configured handle (no @me endpoint)"}
}
