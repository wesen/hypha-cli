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

// --- asks group: hypha asks <list|get|reply|close> ---

// AsksListCommand implements `hypha asks list [--flavor iso|gig] [--status open|closed|all]`.
type AsksListCommand struct{ *cmds.CommandDescription }

type asksListSettings struct {
	ConnectionSettings
	Flavor string `glazed:"flavor"`
	Status string `glazed:"status"`
}

func NewAsksListCommand() (*AsksListCommand, error) {
	desc := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List asks (ISO/gig board) via MCP"),
		cmds.WithFlags(
			fields.New("flavor", fields.TypeString, fields.WithDefault("iso"), fields.WithHelp("iso or gig")),
			fields.New("status", fields.TypeString, fields.WithDefault("open"), fields.WithHelp("open|closed|expired|all")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &AsksListCommand{CommandDescription: desc}, nil
}

func (c *AsksListCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &asksListSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	list, err := client.MCP().ListAsks(ctx, s.Flavor, s.Status)
	if err != nil {
		return err
	}
	for i := range list.Asks {
		a := list.Asks[i]
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("ask_id", a.AskID), types.MRP("flavor", a.Flavor),
			types.MRP("author", a.Author), types.MRP("status", a.Status),
			types.MRP("topics", a.Topics), types.MRP("reply_count", a.ReplyCount),
			types.MRP("last_activity", a.LastActivity), types.MRP("created_at", a.CreatedAt),
		)); err != nil {
			return err
		}
	}
	return nil
}

// AsksGetCommand implements `hypha asks get <id>`.
type AsksGetCommand struct{ *cmds.CommandDescription }

type asksGetSettings struct {
	ConnectionSettings
	ID string `glazed:"id"`
}

func NewAsksGetCommand() (*AsksGetCommand, error) {
	desc := cmds.NewCommandDescription(
		"get",
		cmds.WithShort("Show an ask and its thread (MCP get_ask)"),
		cmds.WithArguments(
			fields.New("id", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("Ask id"), fields.WithIsArgument(true)),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &AsksGetCommand{CommandDescription: desc}, nil
}

func (c *AsksGetCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &asksGetSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	t, err := client.MCP().GetAsk(ctx, s.ID)
	if err != nil {
		return err
	}
	if err := gp.AddRow(ctx, types.NewRow(
		types.MRP("ask_id", t.AskID), types.MRP("flavor", t.Flavor),
		types.MRP("author", t.Author), types.MRP("status", t.Status),
		types.MRP("topics", t.Topics), types.MRP("body", t.Body),
		types.MRP("reply_count", t.ReplyCount),
	)); err != nil {
		return err
	}
	for i := range t.Thread {
		r := t.Thread[i]
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("ask_id", t.AskID), types.MRP("reply_id", r.ID),
			types.MRP("actor", r.Actor), types.MRP("ts", r.Ts),
			types.MRP("body", r.Body),
		)); err != nil {
			return err
		}
	}
	return nil
}

// AsksReplyCommand implements `hypha asks reply <id> <body>`.
type AsksReplyCommand struct{ *cmds.CommandDescription }

type asksReplySettings struct {
	ConnectionSettings
	ID   string `glazed:"id"`
	Body string `glazed:"body"`
}

func NewAsksReplyCommand() (*AsksReplyCommand, error) {
	desc := cmds.NewCommandDescription(
		"reply",
		cmds.WithShort("Reply to an ask thread (MCP reply_ask)"),
		cmds.WithArguments(
			fields.New("id", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("Ask id"), fields.WithIsArgument(true)),
			fields.New("body", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("Reply body"), fields.WithIsArgument(true)),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &AsksReplyCommand{CommandDescription: desc}, nil
}

func (c *AsksReplyCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &asksReplySettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	r, err := client.MCP().ReplyAsk(ctx, s.ID, s.Body)
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("id", r.ID), types.MRP("actor", r.Actor), types.MRP("body", r.Body)))
}

// AsksCloseCommand implements `hypha asks close <id> [--helped <member>] [--note <text>]`.
type AsksCloseCommand struct{ *cmds.CommandDescription }

type asksCloseSettings struct {
	ConnectionSettings
	ID     string `glazed:"id"`
	Helped string `glazed:"helped"`
	Note   string `glazed:"note"`
}

func NewAsksCloseCommand() (*AsksCloseCommand, error) {
	desc := cmds.NewCommandDescription(
		"close",
		cmds.WithShort("Close your own ask (MCP close_ask)"),
		cmds.WithArguments(
			fields.New("id", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("Ask id to close"), fields.WithIsArgument(true)),
		),
		cmds.WithFlags(
			fields.New("helped", fields.TypeString, fields.WithHelp("Member id who helped (records a trust signal)")),
			fields.New("note", fields.TypeString, fields.WithHelp("Closing note")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &AsksCloseCommand{CommandDescription: desc}, nil
}

func (c *AsksCloseCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &asksCloseSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	if err := client.MCP().CloseAsk(ctx, hypha.CloseAskOptions{ID: s.ID, Helped: s.Helped, Note: s.Note}); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("id", s.ID), types.MRP("closed", true)))
}

// --- iso group: hypha iso post ---

// IsoPostCommand implements `hypha iso post <body> [--topic ...]`.
type IsoPostCommand struct{ *cmds.CommandDescription }

type isoPostSettings struct {
	ConnectionSettings
	Body   string   `glazed:"body"`
	Topics []string `glazed:"topic"`
}

func NewIsoPostCommand() (*IsoPostCommand, error) {
	desc := cmds.NewCommandDescription(
		"post",
		cmds.WithShort("Post an ISO to the circle board (MCP post_iso)"),
		cmds.WithArguments(
			fields.New("body", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("What you're looking for"), fields.WithIsArgument(true)),
		),
		cmds.WithFlags(
			fields.New("topic", fields.TypeStringList, fields.WithHelp("Topic channel (repeatable)")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &IsoPostCommand{CommandDescription: desc}, nil
}

func (c *IsoPostCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &isoPostSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	ask, err := client.MCP().PostISO(ctx, hypha.PostISOOptions{Body: s.Body, Topics: s.Topics})
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("ask_id", ask.AskID), types.MRP("status", ask.Status),
		types.MRP("topics", ask.Topics), types.MRP("author", ask.Author),
	))
}
