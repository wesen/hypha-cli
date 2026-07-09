package hypha

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"
)

// Checkpoint is one integrity pin over the event log.
type Checkpoint struct {
	Seq        int    `json:"seq"`
	UpToID     string `json:"up_to_id"`
	EventCount int    `json:"event_count"`
	Hash       string `json:"hash"`
	PrevHash   string `json:"prev_hash"`
	CreatedAt  int64  `json:"created_at"`
}

// Checkpoints lists recent integrity checkpoints (newest first).
// limit defaults to 10, max 200 upstream.
func (c *Client) Checkpoints(ctx context.Context, limit int) ([]Checkpoint, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	var wrap struct {
		Checkpoints []Checkpoint `json:"checkpoints"`
	}
	if err := c.doGet(ctx, "/api/v1/checkpoints", params, &wrap); err != nil {
		return nil, err
	}
	return wrap.Checkpoints, nil
}

// VerifyResult is the outcome of recomputing the integrity chain.
type VerifyResult struct {
	OK           bool   // running == pinnedHash over id <= upToID
	ComputedHash string
	EventCount   int
}

// Verify recomputes the cumulative SHA-256 chain over the caller's events
// (id <= upToID) and compares to pinnedHash. The canonical event format is
// from docs/portability:
//
//	<id>|<ts>|<actor>|<verb>|<target_or_empty>|<value_amount_or_empty>|<value_unit_or_empty>
//
// Empty string for null fields (not the word "null"). body/topics/audience/
// redacted are excluded so redaction does not break the chain.
func (c *Client) Verify(ctx context.Context, upToID, pinnedHash string) (*VerifyResult, error) {
	running := []byte("")
	count := 0
	cursor := ""
	for {
		page, err := c.Export(ctx, cursor)
		if err != nil {
			return nil, err
		}
		for _, ev := range page.Events {
			if ev.ID > upToID {
				return &VerifyResult{OK: hexEqualString(hex.EncodeToString(running), pinnedHash), ComputedHash: hex.EncodeToString(running), EventCount: count}, nil
			}
			running = sha256chain(running, []byte(canonicalEvent(ev)))
			count++
		}
		if page.NextCursor == nil {
			break
		}
		cursor = *page.NextCursor
	}
	return &VerifyResult{OK: hexEqualString(hex.EncodeToString(running), pinnedHash), ComputedHash: hex.EncodeToString(running), EventCount: count}, nil
}

// canonicalEvent formats one event per the portability spec.
func canonicalEvent(ev Event) string {
	return strings.Join([]string{
		ev.ID,
		strconv.FormatInt(ev.Ts, 10),
		ev.Actor,
		ev.Verb,
		orEmpty(ev.Target),
		orEmptyAmount(ev.ValueAmount),
		orEmpty(ev.ValueUnit),
	}, "|")
}

func orEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func orEmptyAmount(p *float64) string {
	if p == nil {
		return ""
	}
	// Match the kernel: integer-valued amounts print without a decimal,
	// fractional print with the minimal decimal form.
	a := *p
	if a == float64(int64(a)) {
		return strconv.FormatInt(int64(a), 10)
	}
	return strconv.FormatFloat(a, 'f', -1, 64)
}

// sha256chain returns sha256(prev + canonical). The chain seeds with "".
func sha256chain(prev, canonical []byte) []byte {
	h := sha256.New()
	h.Write(prev)
	h.Write(canonical)
	return h.Sum(nil)
}

// hexEqualString compares two hex strings case-insensitively.
func hexEqualString(a, b string) bool {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	return a == b
}

