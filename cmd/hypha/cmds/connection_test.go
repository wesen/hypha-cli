package cmds

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/stretchr/testify/require"
)

// TestNewConnectionSection verifies the connection section is built with the
// expected fields and slug.
func TestNewConnectionSection(t *testing.T) {
	s, err := NewConnectionSection()
	require.NoError(t, err)
	require.Equal(t, ConnectionSectionName, s.GetSlug())
}

// TestAddConnectionSections verifies the standard sections attach to a
// command description under their slugs.
func TestAddConnectionSections(t *testing.T) {
	desc := cmds.NewCommandDescription("test")
	require.NoError(t, AddConnectionSections(desc))
	for _, slug := range []string{"glazed", "command-settings", ConnectionSectionName} {
		_, ok := desc.Schema.Get(slug)
		require.True(t, ok, "section %q should be set", slug)
	}
}

// TestStandardSectionsSmoke is a smoke test that all three sections build.
func TestStandardSectionsSmoke(t *testing.T) {
	sections, err := StandardSections()
	require.NoError(t, err)
	require.Len(t, sections, 3)
}

// init sets HYPHA_CONFIG away from the real ~/.config path so any local
// config file does not leak into these tests.
func init() {
	_ = os.Setenv("HYPHA_CONFIG", filepath.Join(os.TempDir(), "hypha-cli-nonexistent-config.json"))
}
