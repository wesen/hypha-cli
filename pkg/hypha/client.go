// Package hypha is a pure-Go client for the Hypha kernel REST API
// (https://hyphahypha.club/docs/readme). It has no Cobra and no goja
// dependency — the Glazed CLI (cmd/hypha) and the go-go-goja module
// (pkg/gojamodules/hypha) both depend on it.
package hypha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client talks to one Hypha kernel deployment. State is just the base URL,
// a PAT (bearer), and an *http.Client.
type Client struct {
	baseURL string
	pat     string
	http    *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the default *http.Client.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.http = c }
}

// WithTimeout sets the underlying http.Client timeout (convenience).
func WithTimeout(d time.Duration) Option {
	return func(cl *Client) { cl.http.Timeout = d }
}

// NewClient returns a client for baseURL with the given PAT (which may be
// empty for unauthenticated calls such as /health). baseURL's trailing
// slash is trimmed.
func NewClient(baseURL, pat string, opts ...Option) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		pat:    pat,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string { return c.baseURL }

// PAT returns the configured PAT (opaque to the client; never inspected).
func (c *Client) PAT() string { return c.pat }

// Error is a typed Hypha API error. The kernel returns {"error":"<msg>"}.
type Error struct {
	Status  int
	Message string
	Body    string
}

func (e *Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("hypha: %d %s", e.Status, e.Message)
	}
	return fmt.Sprintf("hypha: %d", e.Status)
}

// doJSON is the single HTTP helper. in is JSON-encoded as the request body
// (nil for GET/DELETE); out is decoded from the response body (pass nil to
// ignore the body). Query params are added via do's caller through path.
func (c *Client) doJSON(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("hypha: marshal request: %w", err)
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("hypha: build request: %w", err)
	}
	if c.pat != "" {
		req.Header.Set("Authorization", "Bearer "+c.pat)
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("hypha: request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(raw, &e)
		return &Error{Status: resp.StatusCode, Message: e.Error, Body: string(raw)}
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("hypha: decode response: %w (body=%q)", err, string(raw))
		}
	}
	return nil
}

// doGet is a convenience for GET requests with no body, adding query params.
func (c *Client) doGet(ctx context.Context, path string, params url.Values, out any) error {
	full := path
	if params != nil {
		q := params.Encode()
		if q != "" {
			full = path + "?" + q
		}
	}
	return c.doJSON(ctx, http.MethodGet, full, nil, out)
}

// doRaw sends a raw JSON body and returns the raw JSON response. Used by the
// MCP client, whose /mcp endpoint speaks JSON-RPC 2.0 rather than the kernel's
// REST {"error":...} shape.
func (c *Client) doRaw(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("hypha: build request: %w", err)
	}
	if c.pat != "" {
		req.Header.Set("Authorization", "Bearer "+c.pat)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hypha: request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, &Error{Status: resp.StatusCode, Body: string(raw)}
	}
	return raw, nil
}

// Health checks liveness: GET /health (returns bare "ok", not JSON).
func (c *Client) Health(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", &Error{Status: resp.StatusCode, Body: string(b)}
	}
	return strings.TrimSpace(string(b)), nil
}
