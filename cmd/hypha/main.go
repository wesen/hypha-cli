package main

import (
	"os"

	gcmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/hypha-cli/cmd/hypha/cmds"
	"github.com/go-go-golems/hypha-cli/pkg/hypha"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hypha",
	Short: "Hypha kernel CLI (glazed)",
	Long: `hypha is a Glazed CLI client for the Hypha kernel.

It wraps the Hypha REST API (https://hyphahypha.club/docs/readme) and emits
structured output (JSON/YAML/CSV/table) via Glazed. Configure the connection
with --base-url/--pat, env (HYPHA_BASE_URL/HYPHA_PAT), or
~/.config/hypha/config.json (mode 600, override with HYPHA_CONFIG).`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := logging.InitLoggerFromCobra(cmd); err != nil {
			return err
		}
		return loadHyphaConfig()
	},
}

// parserCfg is the shared Glazed parser config (env prefix HYPHA).
var parserCfg = cli.CobraParserConfig{AppName: "hypha"}

// build wraps a Glazed command in a Cobra command.
func build(c gcmds.Command) (*cobra.Command, error) {
	return cli.BuildCobraCommandFromCommand(c,
		cli.WithParserConfig(parserCfg),
	)
}

func main() {
	if err := logging.AddLoggingSectionToRootCommand(rootCmd, "hypha"); err != nil {
		cobra.CheckErr(err)
	}

	helpSystem := help.NewHelpSystem()
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	// Identity + events
	cobra.CheckErr(addCmd(rootCmd, mustBuild(cmds.NewPostCommand)))
	cobra.CheckErr(addCmd(rootCmd, mustBuild(cmds.NewFeedCommand)))
	cobra.CheckErr(addCmd(rootCmd, mustBuild(cmds.NewSearchCommand)))
	cobra.CheckErr(addCmd(rootCmd, mustBuild(cmds.NewRedactCommand)))
	cobra.CheckErr(addCmd(rootCmd, mustBuild(cmds.NewMembersCommand)))
	cobra.CheckErr(addCmd(rootCmd, mustBuild(cmds.NewWhoisCommand)))
	cobra.CheckErr(addCmd(rootCmd, mustBuild(cmds.NewWhoamiCommand)))

	// Views group
	views := &cobra.Command{Use: "views", Short: "Read-side views over the event log"}
	rootCmd.AddCommand(views)
	cobra.CheckErr(addCmd(views, mustBuild(cmds.NewBalanceCommand)))
	cobra.CheckErr(addCmd(views, mustBuild(cmds.NewTrustCommand)))
	cobra.CheckErr(addCmd(views, mustBuild(cmds.NewTopicsCommand)))
	cobra.CheckErr(addCmd(views, mustBuild(cmds.NewWhoCanHelpCommand)))
	cobra.CheckErr(addCmd(views, mustBuild(cmds.NewGratitudeCommand)))

	// PATs group
	pats := &cobra.Command{Use: "pats", Short: "Manage Personal Access Tokens"}
	rootCmd.AddCommand(pats)
	cobra.CheckErr(addCmd(pats, mustBuild(cmds.NewPatsMintCommand)))
	cobra.CheckErr(addCmd(pats, mustBuild(cmds.NewPatsListCommand)))
	cobra.CheckErr(addCmd(pats, mustBuild(cmds.NewPatsRevokeCommand)))

	// Webhooks group
	webhooks := &cobra.Command{Use: "webhooks", Short: "Manage webhook deliveries"}
	rootCmd.AddCommand(webhooks)
	cobra.CheckErr(addCmd(webhooks, mustBuild(cmds.NewWebhookSubscribeCommand)))
	cobra.CheckErr(addCmd(webhooks, mustBuild(cmds.NewWebhooksListCommand)))
	cobra.CheckErr(addCmd(webhooks, mustBuild(cmds.NewWebhookUnsubscribeCommand)))

	// Portability + integrity
	cobra.CheckErr(addCmd(rootCmd, mustBuild(cmds.NewExportCommand)))
	cobra.CheckErr(addCmd(rootCmd, mustBuild(cmds.NewCheckpointsCommand)))
	cobra.CheckErr(addCmd(rootCmd, mustBuild(cmds.NewVerifyCommand)))

	// Asks (ISO/gig board) via MCP
	asks := &cobra.Command{Use: "asks", Short: "Ask board (ISO/gig) via MCP: list, get, reply, close"}
	rootCmd.AddCommand(asks)
	cobra.CheckErr(addCmd(asks, mustBuild(cmds.NewAsksListCommand)))
	cobra.CheckErr(addCmd(asks, mustBuild(cmds.NewAsksGetCommand)))
	cobra.CheckErr(addCmd(asks, mustBuild(cmds.NewAsksReplyCommand)))
	cobra.CheckErr(addCmd(asks, mustBuild(cmds.NewAsksCloseCommand)))

	iso := &cobra.Command{Use: "iso", Short: "Post an ISO to the circle board (MCP)"}
	rootCmd.AddCommand(iso)
	cobra.CheckErr(addCmd(iso, mustBuild(cmds.NewIsoPostCommand)))

	cobra.CheckErr(rootCmd.Execute())
}

// mustBuild constructs a Glazed command and panics on construction error.
// (Construction errors are programming errors, not runtime errors.)
func mustBuild[T any](newCmd func() (*T, error)) *T {
	c, err := newCmd()
	if err != nil {
		panic(err)
	}
	return c
}

// addCmd builds a Glazed command value into a Cobra command and attaches it
// to parent.
func addCmd(parent *cobra.Command, c gcmds.Command) error {
	cc, err := build(c)
	if err != nil {
		return err
	}
	parent.AddCommand(cc)
	return nil
}

// loadHyphaConfig reads ~/.config/hypha/config.json (or $HYPHA_CONFIG) and seeds
// the HYPHA_* env vars for any field not already set. The connection section's
// AppName=hypha env source then picks them up as defaults, so the precedence is:
// explicit --flags > env (already set) > config file > built-in defaults.
// A missing or malformed config file is logged but not fatal.
func loadHyphaConfig() error {
	cfg, err := hypha.LoadConfig()
	if err != nil {
		// Malformed config is not fatal; flags/env can still supply values.
		return nil
	}
	if cfg.BaseURL != "" && os.Getenv("HYPHA_BASE_URL") == "" {
		_ = os.Setenv("HYPHA_BASE_URL", cfg.BaseURL)
	}
	if cfg.PAT != "" && os.Getenv("HYPHA_PAT") == "" {
		_ = os.Setenv("HYPHA_PAT", cfg.PAT)
	}
	if cfg.Handle != "" && os.Getenv("HYPHA_HANDLE") == "" {
		_ = os.Setenv("HYPHA_HANDLE", cfg.Handle)
	}
	return nil
}
