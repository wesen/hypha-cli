// Package doc embeds the hypha JS module API help pages served by the
// xgoja provider's HelpSource.
package doc

import "embed"

//go:embed *.md
var helpFS embed.FS

// FS returns the embedded help filesystem (root = ".").
func FS() embed.FS { return helpFS }
