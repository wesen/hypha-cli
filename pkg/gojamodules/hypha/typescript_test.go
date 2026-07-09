package hyphamod_test

import (
	"strings"
	"testing"

	"github.com/go-go-golems/go-go-goja/modules"
	"github.com/stretchr/testify/require"
)

// TestModuleIsTypeScriptDeclarer confirms the module implements the optional
// declarer interface so xgoja gen-dts emits a hypha.d.ts.
func TestModuleIsTypeScriptDeclarer(t *testing.T) {
	mod := modules.GetModule("hypha")
	require.NotNil(t, mod)
	d, ok := mod.(modules.TypeScriptDeclarer)
	require.True(t, ok, "module must implement TypeScriptDeclarer")
	ts := d.TypeScriptModule()
	require.Equal(t, "hypha", ts.Name)
	require.NotEmpty(t, ts.RawDTS)
}

// TestTypeScriptModuleShape sanity-checks the generated declarations contain
// the key PascalCase interfaces so the .d.ts tracks the real Go API.
func TestTypeScriptModuleShape(t *testing.T) {
	d, ok := modules.GetModule("hypha").(modules.TypeScriptDeclarer)
	require.True(t, ok)
	joined := strings.Join(d.TypeScriptModule().RawDTS, "\n")
	for _, want := range []string{
		"export interface EventBuilder",
		"Post(): Event",
		"postISO",
		"AskThread",
		"ValidationResult",
		"verify(): VerifyQuery",
	} {
		require.Contains(t, joined, want, "d.ts should declare %q", want)
	}
}
