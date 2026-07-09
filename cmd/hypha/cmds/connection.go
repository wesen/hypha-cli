// Package cmds implements the Glazed CLI verbs for the `hypha` binary.
package cmds

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/settings"
)

// ConnectionSectionName is the custom Glazed section carrying --base-url and
// --pat so every verb shares the same connection flags without redefining them.
const ConnectionSectionName = "connection"

// NewConnectionSection returns the shared connection section. Defaults come
// from the loaded config file (set on the root) or env (HYPHA_BASE_URL,
// HYPHA_PAT) via the AppName env source.
func NewConnectionSection() (schema.Section, error) {
	return schema.NewSection(ConnectionSectionName, "Hypha connection",
		schema.WithFields(
			fields.New("base-url", fields.TypeString,
				fields.WithDefault("http://localhost:8787"),
				fields.WithHelp("Hypha base URL (e.g. https://hyphahypha.club)")),
			fields.New("pat", fields.TypeString,
				fields.WithHelp("Personal Access Token (hh_pat_...); overrides config file")),
		),
	)
}

// ConnectionSettings holds the decoded connection flags.
type ConnectionSettings struct {
	BaseURL string `glazed:"base-url"`
	PAT     string `glazed:"pat"`
}

// StandardSections returns the glazed + command-settings + connection sections
// for use with cmds.WithSections(...). Every verb calls this in its constructor.
func StandardSections() ([]schema.Section, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	cmdSettings, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	conn, err := NewConnectionSection()
	if err != nil {
		return nil, err
	}
	return []schema.Section{glazedSection, cmdSettings, conn}, nil
}

// WithStandardSections returns a cmds.CommandDescriptionOption that attaches
// the glazed + command-settings + connection sections. It panics if section
// construction fails (use StandardSections() if you need the error).
func WithStandardSections() cmds.CommandDescriptionOption {
	sections, err := StandardSections()
	if err != nil {
		panic(err)
	}
	return cmds.WithSections(sections...)
}

// AddConnectionSections attaches the standard section set to an already-built
// CommandDescription (for constructors that create the description first).
func AddConnectionSections(desc *cmds.CommandDescription) error {
	sections, err := StandardSections()
	if err != nil {
		return err
	}
	for _, s := range sections {
		desc.Schema.Set(s.GetSlug(), s)
	}
	return nil
}
