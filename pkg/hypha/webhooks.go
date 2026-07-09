package hypha

import (
	"context"
	"net/http"
	"net/url"
)

// Webhook is a registered webhook delivery target.
type Webhook struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	CreatedAt int64  `json:"created_at"`
}

// SubscribeOptions is the body of POST /api/v1/webhooks.
type SubscribeOptions struct {
	URL    string `json:"url"`
	Secret string `json:"secret,omitempty"`
}

// Subscribe registers a webhook (HMAC-signed, SSRF-guarded upstream).
func (c *Client) Subscribe(ctx context.Context, o SubscribeOptions) (*Webhook, error) {
	var wh Webhook
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/webhooks", o, &wh); err != nil {
		return nil, err
	}
	return &wh, nil
}

// Webhooks lists registered webhooks (GET /api/v1/webhooks). Envelope:
// {"webhooks":[...]}.
func (c *Client) Webhooks(ctx context.Context) ([]Webhook, error) {
	var wrap struct {
		Webhooks []Webhook `json:"webhooks"`
	}
	if err := c.doGet(ctx, "/api/v1/webhooks", nil, &wrap); err != nil {
		return nil, err
	}
	return wrap.Webhooks, nil
}

// Unsubscribe deletes a webhook (DELETE /api/v1/webhooks/:id).
func (c *Client) Unsubscribe(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/api/v1/webhooks/"+url.PathEscape(id), nil, nil)
}
