package hypha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// InviteOptions seeds an email invite link. The admin endpoint uses
// X-Admin-Token; the member endpoint uses PAT bearer.
type InviteOptions struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Note  string `json:"note,omitempty"`
}

// Invite creates an invite link (POST /api/v1/invites). Returns the invite
// URL (the recipient extracts the token after '#').
func (c *Client) Invite(ctx context.Context, o InviteOptions) (string, error) {
	var raw string
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/invites", o, &raw); err != nil {
		return "", err
	}
	return raw, nil
}

// InviteAdmin creates an invite via the admin endpoint
// (POST /admin/invites) using X-Admin-Token instead of a PAT. Pass the admin
// token; this call bypasses the client's PAT.
func (c *Client) InviteAdmin(ctx context.Context, adminToken string, o InviteOptions) (string, error) {
	body, err := json.Marshal(o)
	if err != nil {
		return "", fmt.Errorf("hypha: marshal invite: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/admin/invites", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Admin-Token", adminToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", &Error{Status: resp.StatusCode, Body: string(raw)}
	}
	return string(raw), nil
}

// AcceptInviteOptions is the body of POST /api/v1/invites/accept.
type AcceptInviteOptions struct {
	Token  string `json:"token"`
	Handle string `json:"handle"`
}

// AcceptInviteResult holds the minted root PAT and member id.
type AcceptInviteResult struct {
	Token   string `json:"token"`   // the minted hh_pat_... PAT (shown once)
	ID      string `json:"id"`      // new member id, when present
	Handle  string `json:"handle"`  // confirmed handle, when present
}

// AcceptInvite accepts an invite and mints the root PAT
// (POST /api/v1/invites/accept). The response shape is not fully documented;
// we decode known fields and fall back to scanning the body for a hh_pat_
// token via ExtractPAT.
func (c *Client) AcceptInvite(ctx context.Context, o AcceptInviteOptions) (*AcceptInviteResult, error) {
	var raw string
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/invites/accept", o, &raw); err != nil {
		return nil, err
	}
	res := &AcceptInviteResult{}
	// Try structured decode first.
	var structured struct {
		Token  string `json:"token"`
		ID     string `json:"id"`
		Handle string `json:"handle"`
	}
	if json.Unmarshal([]byte(raw), &structured) == nil && structured.Token != "" {
		res.Token, res.ID, res.Handle = structured.Token, structured.ID, structured.Handle
	} else {
		res.Token = ExtractPAT(raw)
	}
	return res, nil
}

// ConnectOptions introduces two members (POST /api/v1/connections).
type ConnectOptions struct {
	A    string `json:"a"`         // member id or @handle
	B    string `json:"b"`         // member id or @handle
	Note string `json:"note,omitempty"`
}

// Connect introduces two members; the kernel emails both and appends a
// connect event.
func (c *Client) Connect(ctx context.Context, o ConnectOptions) error {
	var ok struct {
		OK bool `json:"ok"`
	}
	return c.doJSON(ctx, http.MethodPost, "/api/v1/connections", o, &ok)
}

// ExtractPAT scans an arbitrary string for a Hypha PAT (hh_pat_<hex>) and
// returns it, or "" if none found.
func ExtractPAT(s string) string {
	i := strings.Index(s, "hh_pat_")
	if i < 0 {
		return ""
	}
	j := i + len("hh_pat_")
	for j < len(s) && isHex(s[j]) {
		j++
	}
	return s[i:j]
}

func isHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f')
}
