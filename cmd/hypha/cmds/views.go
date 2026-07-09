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

// ViewsGroup implements `hypha views` (the read-side lenses).
type ViewsGroup struct {
	*cmds.CommandDescription
}

type viewSettings struct {
	ConnectionSettings
	Topic  string `glazed:"topic"`
	Member string `glazed:"member"`
}

func NewViewsGroup() (*ViewsGroup, error) {
	desc := cmds.NewCommandDescription(
		"views",
		cmds.WithShort("Read-side views over the event log (balance, trust, topics, ...)"),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &ViewsGroup{CommandDescription: desc}, nil
}

// BalanceCommand implements `hypha views balance`.
type BalanceCommand struct{ *cmds.CommandDescription }

func NewBalanceCommand() (*BalanceCommand, error) {
	desc := cmds.NewCommandDescription("balance", cmds.WithShort("Time balances for all members"))
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &BalanceCommand{CommandDescription: desc}, nil
}

func (c *BalanceCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &viewSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	rows, err := client.Balance(ctx)
	if err != nil {
		return err
	}
	for i := range rows {
		r := rows[i]
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("id", r.ID), types.MRP("handle", r.Handle), types.MRP("name", r.Name),
			types.MRP("received", r.Received), types.MRP("given", r.Given), types.MRP("balance", r.Balance),
		)); err != nil {
			return err
		}
	}
	return nil
}

// TrustCommand implements `hypha views trust`.
type TrustCommand struct{ *cmds.CommandDescription }

func NewTrustCommand() (*TrustCommand, error) {
	desc := cmds.NewCommandDescription("trust", cmds.WithShort("Your trust-graph edges"))
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &TrustCommand{CommandDescription: desc}, nil
}

func (c *TrustCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &viewSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	res, err := client.Trust(ctx)
	if err != nil {
		return err
	}
	for i := range res.Edges {
		e := res.Edges[i]
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("member", res.Member), types.MRP("peer", e.Peer),
			types.MRP("type", e.Type), types.MRP("weight", e.Weight), types.MRP("last_ts", e.LastTs),
		)); err != nil {
			return err
		}
	}
	return nil
}

// TopicsCommand implements `hypha views topics`.
type TopicsCommand struct{ *cmds.CommandDescription }

func NewTopicsCommand() (*TopicsCommand, error) {
	desc := cmds.NewCommandDescription("topics", cmds.WithShort("Active topics in the circle"))
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &TopicsCommand{CommandDescription: desc}, nil
}

func (c *TopicsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &viewSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	rows, err := client.Topics(ctx)
	if err != nil {
		return err
	}
	for i := range rows {
		r := rows[i]
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("topic", r.Topic), types.MRP("events", r.Events),
			types.MRP("actors", r.Actors), types.MRP("last_event_id", r.LastEventID),
		)); err != nil {
			return err
		}
	}
	return nil
}

// WhoCanHelpCommand implements `hypha views who-can-help --topic <t>`.
type WhoCanHelpCommand struct{ *cmds.CommandDescription }

func NewWhoCanHelpCommand() (*WhoCanHelpCommand, error) {
	desc := cmds.NewCommandDescription(
		"who-can-help",
		cmds.WithShort("Members ranked by helpfulness on a topic"),
		cmds.WithFlags(
			fields.New("topic", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Topic to query")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &WhoCanHelpCommand{CommandDescription: desc}, nil
}

func (c *WhoCanHelpCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &viewSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	rows, err := client.WhoCanHelp(ctx, s.Topic)
	if err != nil {
		return err
	}
	for i := range rows {
		r := rows[i]
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("member", r.Member), types.MRP("posts", r.Posts),
			types.MRP("hours_given", r.HoursGiven), types.MRP("balance", r.Balance),
			types.MRP("score", r.Score),
		)); err != nil {
			return err
		}
	}
	return nil
}

// GratitudeCommand implements `hypha views gratitude [@x]`.
type GratitudeCommand struct{ *cmds.CommandDescription }

func NewGratitudeCommand() (*GratitudeCommand, error) {
	desc := cmds.NewCommandDescription(
		"gratitude",
		cmds.WithShort("Kudos-based gratitude scores"),
		cmds.WithFlags(
			fields.New("member", fields.TypeString, fields.WithHelp("Member id or @handle (omit for all)")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &GratitudeCommand{CommandDescription: desc}, nil
}

func (c *GratitudeCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &viewSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	res, err := client.Gratitude(ctx, s.Member)
	if err != nil {
		return err
	}
	for i := range res.Members {
		r := res.Members[i]
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("member", r.Member), types.MRP("handle", r.Handle),
			types.MRP("kudos_received", r.KudosReceived), types.MRP("gratitude", r.Gratitude),
		)); err != nil {
			return err
		}
	}
	return nil
}
