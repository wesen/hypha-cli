package hyphamod_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-go-golems/hypha-cli/pkg/hypha"
	"github.com/stretchr/testify/require"
)

// mcpResult builds a JSON-RPC response whose content[0].text is the JSON
// encoding of v (the real MCP envelope shape).
func mcpResult(t *testing.T, v any) string {
	t.Helper()
	inner, err := json.Marshal(v)
	require.NoError(t, err)
	resp := map[string]any{
		"jsonrpc": "2.0", "id": 1,
		"result": map[string]any{
			"content": []map[string]any{{"type": "text", "text": string(inner)}},
		},
	}
	b, err := json.Marshal(resp)
	require.NoError(t, err)
	return string(b)
}

func mcpServer(t *testing.T, fn func(w http.ResponseWriter, r *http.Request)) (*hypha.Client, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", fn)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return hypha.NewClient(srv.URL, "hh_pat_test"), srv
}

func TestMCPCallToolDecodesEnvelope(t *testing.T) {
	body := mcpResult(t, map[string]any{"asks": []map[string]any{
		{"ask_id": "01A", "flavor": "iso", "status": "open", "topics": []string{"xgoja"}},
	}})
	c, _ := mcpServer(t, func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, body) })
	list, err := c.MCP().ListAsks(context.Background(), "iso", "open")
	require.NoError(t, err)
	require.Len(t, list.Asks, 1)
	require.Equal(t, "01A", list.Asks[0].AskID)
	require.Equal(t, "xgoja", list.Asks[0].Topics[0])
}

func TestMCPCallToolPropagatesProtocolError(t *testing.T) {
	c, _ := mcpServer(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32003,"message":"missing scope"}}`)
	})
	_, err := c.MCP().ListAsks(context.Background(), "iso", "open")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing scope")
}

func TestMCPCallToolDetectsIsError(t *testing.T) {
	resp := map[string]any{
		"jsonrpc": "2.0", "id": 1,
		"result": map[string]any{
			"content": []map[string]any{{"type": "text", "text": "Error: iso not found", "isError": true}},
		},
	}
	b, _ := json.Marshal(resp)
	c, _ := mcpServer(t, func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, string(b)) })
	_, err := c.MCP().GetAsk(context.Background(), "nope")
	require.Error(t, err)
	require.Contains(t, err.Error(), "iso not found")
}

func TestMCPPostISO(t *testing.T) {
	body := mcpResult(t, map[string]any{
		"ask_id": "01NEW", "flavor": "iso", "status": "open", "topics": []string{"go"}, "author": "01ME",
	})
	c, _ := mcpServer(t, func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, body) })
	ask, err := c.MCP().PostISO(context.Background(), hypha.PostISOOptions{Body: "help", Topics: []string{"go"}})
	require.NoError(t, err)
	require.Equal(t, "01NEW", ask.AskID)
	require.Equal(t, "open", ask.Status)
}
