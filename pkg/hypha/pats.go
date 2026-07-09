package hypha

import (
	"context"
	"net/http"
)

// PAT is a Personal Access Token record. The token value is shown only at
// mint time; list responses omit it.
type PAT struct {
	ID        string   `json:"id"`
	Scopes    []string `json:"scopes"`
	Label     string   `json:"label"`
	CreatedAt int64    `json:"created_at"`
	RevokedAt *int64   `json:"revoked_at"`
}

// MintPATOptions is the body of POST /api/v1/pats. Requested scopes must be
// a subset of the presenting credential's scopes (enforced upstream).
type MintPATOptions struct {
	Scopes []string `json:"scopes"`
	Label  string   `json:"label,omitempty"`
}

// MintPATResult is the mint response. Token is the hh_pat_... value (shown
// once).
type MintPATResult struct {
	Token string `json:"token"`
	ID    string `json:"id"`
}

// MintPAT mints a narrower PAT (POST /api/v1/pats).
func (c *Client) MintPAT(ctx context.Context, o MintPATOptions) (*MintPATResult, error) {
	var res MintPATResult
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/pats", o, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// ListPATs lists the caller's credentials (GET /api/v1/pats). Envelope:
// {"pats":[...]}. Tokens are never re-shown.
func (c *Client) ListPATs(ctx context.Context) ([]PAT, error) {
	var wrap struct {
		Pats []PAT `json:"pats"`
	}
	if err := c.doGet(ctx, "/api/v1/pats", nil, &wrap); err != nil {
		return nil, err
	}
	return wrap.Pats, nil
}

// RevokePAT revokes a credential immediately (DELETE /api/v1/pats/:id).
func (c *Client) RevokePAT(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/api/v1/pats/"+id, nil, nil)
}
