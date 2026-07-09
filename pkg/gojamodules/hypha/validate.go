package hyphamod

import "strings"

// reservedBareKinds is the M1 first-party reserved kind namespace (no dot).
// Third-party tools MUST dot-prefix their kinds (docs/conventions).
var reservedBareKinds = map[string]bool{
	"iso": true, "gig": true, "meeting": true, "update": true, "chat": true,
	"request": true, "accept": true, "decline": true, "help": true, "close": true,
	"kudos": true, "tip": true,
}

// ValidationResult is a Go-backed result object (reflected to JS like
// markdown's ValidationResult).
type ValidationResult struct {
	Valid  bool
	Errors []string
}

// validateKind appends an error if kind is non-empty, not dot-prefixed, and not
// a reserved bare kind.
func validateKind(kind string) (errs []string) {
	if kind == "" {
		return nil
	}
	if strings.Contains(kind, ".") {
		return nil // dot-prefixed third-party kind is valid
	}
	if reservedBareKinds[kind] {
		return nil
	}
	return []string{"kind " + kind + " is not a reserved bare kind; third-party kinds must be dot-prefixed (e.g. \"acme.bounty\")"}
}
