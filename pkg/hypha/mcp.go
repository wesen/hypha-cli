package hypha

import (
	"context"
	"encoding/json"
	"fmt"
)

// MCPClient calls the Hypha MCP endpoint (POST /mcp) with JSON-RPC 2.0.
// It reuses the parent Client's base URL, PAT, and http.Client. The MCP
// surface is the only way to reach the ask (ISO/gig) domain — post_iso,
// reply_ask, close_ask, list_asks, get_ask — which lives in a separate store
// from the raw /api/v1/events log.
type MCPClient struct {
	c *Client
}

// MCP returns a view of the client that calls MCP tools.
func (c *Client) MCP() *MCPClient { return &MCPClient{c: c} }

// mcpRequest is a JSON-RPC 2.0 request.
type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// mcpResponse is a JSON-RPC 2.0 response. On success, Result is non-nil; on
// protocol error, Err is non-nil.
type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *mcpError) Error() string { return fmt.Sprintf("mcp: %d %s", e.Code, e.Message) }

// mcpContent is the MCP result envelope: an array of content blocks whose
// first text block holds the real data as a JSON string.
type mcpContent struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// call invokes one MCP tool and unmarshals the inner JSON payload (the text
// block's text) into out. If the tool itself failed, the text is "Error: ..."
// and isError is true; call detects that and returns an error.
func (m *MCPClient) call(ctx context.Context, tool string, args any, out any) error {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("hypha.mcp: marshal args: %w", err)
	}
	params, _ := json.Marshal(map[string]any{"name": tool, "arguments": json.RawMessage(argsJSON)})
	reqBody, _ := json.Marshal(mcpRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: params})

	// Use the raw HTTP path (not doJSON) because /mcp is JSON-RPC, not the
	// kernel's {"error":...} REST shape.
	resp, err := m.c.doRaw(ctx, "POST", "/mcp", reqBody)
	if err != nil {
		return err
	}
	var mr mcpResponse
	if err := json.Unmarshal(resp, &mr); err != nil {
		return fmt.Errorf("hypha.mcp: decode response: %w (body=%q)", err, string(resp))
	}
	if mr.Error != nil {
		return mr.Error
	}
	// Decode the content envelope.
	var env struct {
		Content []struct {
			Type    string `json:"type"`
			Text    string `json:"text"`
			IsError bool   `json:"isError,omitempty"`
		} `json:"content"`
	}
	if err := json.Unmarshal(mr.Result, &env); err != nil {
		return fmt.Errorf("hypha.mcp: decode content: %w", err)
	}
	if len(env.Content) == 0 {
		return fmt.Errorf("hypha.mcp: empty content")
	}
	text := env.Content[0].Text
	if env.Content[0].IsError || (len(text) > 7 && text[:7] == "Error: ") {
		return fmt.Errorf("hypha.mcp: %s", text)
	}
	if out != nil && len(text) > 0 {
		if err := json.Unmarshal([]byte(text), out); err != nil {
			// Some tools return a bare string like "\"ok\"" or a non-JSON
			// confirmation; treat a decode failure as a non-fatal result.
			if s, ok := out.(*string); ok {
				*s = text
				return nil
			}
		}
	}
	return nil
}

// Ask is one row from list_asks (an ISO or gig). Fields mirror the MCP shape.
type Ask struct {
	AskID        string   `json:"ask_id"`
	RootEventID  string   `json:"root_event_id"`
	Flavor       string   `json:"flavor"` // "iso" | "gig"
	GigType      *string  `json:"gig_type"`
	Author       string   `json:"author"`
	Title        *string  `json:"title"`
	Reward       *string  `json:"reward"`
	Timeframe    *string  `json:"timeframe"`
	Client       *string  `json:"client"`
	Topics       []string `json:"topics"`
	Status       string   `json:"status"`
	ReplyCount   int      `json:"reply_count"`
	LastActivity int64    `json:"last_activity"`
	CreatedAt    int64    `json:"created_at"`
	ClosedTarget  *string  `json:"closed_target"`
}

// AskList is the list_asks payload.
type AskList struct {
	Asks []Ask `json:"asks"`
}

// ListAsks calls the list_asks MCP tool. flavor is "iso" or "gig"; status is
// "open" (default), "closed", "expired", or "all".
func (m *MCPClient) ListAsks(ctx context.Context, flavor, status string) (*AskList, error) {
	args := map[string]any{}
	if flavor != "" {
		args["flavor"] = flavor
	}
	if status != "" {
		args["status"] = status
	}
	var out AskList
	if err := m.call(ctx, "list_asks", args, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AskThread is the get_ask payload: the flat Ask fields plus a top-level
// body and a thread of reply events. The real MCP shape is a flat ask with
// "body" and "thread" at the top level (no nested "ask" object), so AskThread
// embeds Ask to capture all the ask fields.
type AskThread struct {
	Ask    // flat ask fields (AskID, Status, Topics, Author, ...)
	Body   string     `json:"body"`
	Thread []AskReply `json:"thread"`
}

// AskReply is one reply event in an ask thread. It mirrors the event shape.
type AskReply struct {
	ID        string   `json:"id"`
	Ts        int64    `json:"ts"`
	Actor     string   `json:"actor"`
	Verb      string   `json:"verb"`
	Kind      *string  `json:"kind"`
	Target    *string  `json:"target"`
	Ref       *string  `json:"ref"`
	Topics   []string `json:"topics"`
	Audience string   `json:"audience"`
	Body     string   `json:"body"`
	Redacted int      `json:"redacted"`
}

// GetAsk calls the get_ask MCP tool (an ask plus its replies).
func (m *MCPClient) GetAsk(ctx context.Context, id string) (*AskThread, error) {
	var out AskThread
	if err := m.call(ctx, "get_ask", map[string]string{"id": id}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PostISOOptions is the post_iso tool input.
type PostISOOptions struct {
	Body   string   `json:"body"`
	Topics []string `json:"topics,omitempty"`
}

// PostISO calls post_iso and returns the created ask.
func (m *MCPClient) PostISO(ctx context.Context, o PostISOOptions) (*Ask, error) {
	var out Ask
	if err := m.call(ctx, "post_iso", o, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReplyAsk calls reply_ask on an ask thread.
func (m *MCPClient) ReplyAsk(ctx context.Context, id, body string) (*AskReply, error) {
	var out AskReply
	if err := m.call(ctx, "reply_ask", map[string]string{"id": id, "body": body}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CloseAskOptions is the close_ask tool input.
type CloseAskOptions struct {
	ID     string `json:"id"`
	Helped  string `json:"helped,omitempty"`  // member id who helped
	Note   string `json:"note,omitempty"`
}

// CloseAsk calls close_ask on your own ask.
func (m *MCPClient) CloseAsk(ctx context.Context, o CloseAskOptions) error {
	var s string
	if err := m.call(ctx, "close_ask", o, &s); err != nil {
		return err
	}
	return nil
}
