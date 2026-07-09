package cmds

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/hypha-cli/pkg/hypha"
)

// ExportCommand implements `hypha export [--cursor]`. With no cursor it
// follows every page (ExportAll); with --cursor it fetches one page.
type ExportCommand struct{ *cmds.CommandDescription }

type exportSettings struct {
	ConnectionSettings
	Cursor string `glazed:"cursor"`
}

func NewExportCommand() (*ExportCommand, error) {
	desc := cmds.NewCommandDescription(
		"export",
		cmds.WithShort("Export your full event history (paginated; follows all pages by default)"),
		cmds.WithFlags(
			fields.New("cursor", fields.TypeString, fields.WithHelp("Resume from a cursor (fetch one page)")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &ExportCommand{CommandDescription: desc}, nil
}

func (c *ExportCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &exportSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)

	if s.Cursor != "" {
		page, err := client.Export(ctx, s.Cursor)
		if err != nil {
			return err
		}
		for i := range page.Events {
			if err := gp.AddRow(ctx, eventRow(&page.Events[i])); err != nil {
				return err
			}
		}
		return nil
	}

	ex, err := client.ExportAll(ctx)
	if err != nil {
		return err
	}
	for i := range ex.Events {
		if err := gp.AddRow(ctx, eventRow(&ex.Events[i])); err != nil {
			return err
		}
	}
	return nil
}

// CheckpointsCommand implements `hypha checkpoints [--limit]`.
type CheckpointsCommand struct{ *cmds.CommandDescription }

type checkpointsSettings struct {
	ConnectionSettings
	Limit int `glazed:"limit"`
}

func NewCheckpointsCommand() (*CheckpointsCommand, error) {
	desc := cmds.NewCommandDescription(
		"checkpoints",
		cmds.WithShort("List recent integrity checkpoints"),
		cmds.WithFlags(
			fields.New("limit", fields.TypeInteger, fields.WithHelp("Max checkpoints (default 10, max 200)")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &CheckpointsCommand{CommandDescription: desc}, nil
}

func (c *CheckpointsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &checkpointsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	cps, err := client.Checkpoints(ctx, s.Limit)
	if err != nil {
		return err
	}
	for i := range cps {
		cp := cps[i]
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("seq", cp.Seq), types.MRP("up_to_id", cp.UpToID),
			types.MRP("event_count", cp.EventCount), types.MRP("hash", cp.Hash),
			types.MRP("prev_hash", cp.PrevHash), types.MRP("created_at", cp.CreatedAt),
		)); err != nil {
			return err
		}
	}
	return nil
}

// VerifyCommand implements `hypha verify --up-to <id> --hash <h>`.
type VerifyCommand struct{ *cmds.CommandDescription }

type verifySettings struct {
	ConnectionSettings
	UpTo string `glazed:"up-to"`
	Hash string `glazed:"hash"`
}

func NewVerifyCommand() (*VerifyCommand, error) {
	desc := cmds.NewCommandDescription(
		"verify",
		cmds.WithShort("Recompute the integrity chain over your events and compare to a pinned hash"),
		cmds.WithFlags(
			fields.New("up-to", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Verify events with id <= this (checkpoint up_to_id)")),
			fields.New("hash", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Pinned cumulative SHA-256 hash")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &VerifyCommand{CommandDescription: desc}, nil
}

func (c *VerifyCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &verifySettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	res, err := client.Verify(ctx, s.UpTo, s.Hash)
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("ok", res.OK), types.MRP("computed_hash", res.ComputedHash),
		types.MRP("event_count", res.EventCount),
	))
}
