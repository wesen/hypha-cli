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

func memberRow(m *hypha.Member) types.Row {
	return types.NewRow(
		types.MRP("id", m.ID),
		types.MRP("handle", m.Handle),
		types.MRP("name", m.Name),
		types.MRP("bio", m.Bio),
		types.MRP("timezone", m.Timezone),
		types.MRP("status", m.Status),
		types.MRP("invited_by", m.InvitedBy),
		types.MRP("created_at", m.CreatedAt),
	)
}

// MembersCommand implements `hypha members`.
type MembersCommand struct{ *cmds.CommandDescription }

type MembersSettings struct{ ConnectionSettings }

func NewMembersCommand() (*MembersCommand, error) {
	desc := cmds.NewCommandDescription("members", cmds.WithShort("List circle members"))
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &MembersCommand{CommandDescription: desc}, nil
}

func (c *MembersCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &MembersSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	ms, err := client.Members(ctx)
	if err != nil {
		return err
	}
	for i := range ms {
		if err := gp.AddRow(ctx, memberRow(&ms[i])); err != nil {
			return err
		}
	}
	return nil
}

// WhoisCommand implements `hypha whois <@x>`.
type WhoisCommand struct{ *cmds.CommandDescription }

type WhoisSettings struct {
	ConnectionSettings
	Handle string `glazed:"handle"`
}

func NewWhoisCommand() (*WhoisCommand, error) {
	desc := cmds.NewCommandDescription(
		"whois",
		cmds.WithShort("Identity + trust summary for a member"),
		cmds.WithArguments(
			fields.New("handle", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("Member id or @handle"), fields.WithIsArgument(true)),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &WhoisCommand{CommandDescription: desc}, nil
}

func (c *WhoisCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &WhoisSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	m, err := client.Whois(ctx, s.Handle)
	if err != nil {
		return err
	}
	row := memberRow(m)
	if m.Trust != nil {
		row.Set("balance", m.Trust.Balance)
		row.Set("hours_given", m.Trust.HoursGiven)
		row.Set("hours_received", m.Trust.HoursReceived)
		row.Set("invites_given", m.Trust.InvitesGiven)
		row.Set("connections", m.Trust.Connections)
	}
	return gp.AddRow(ctx, row)
}

// WhoamiCommand implements `hypha whoami` (resolves via config handle).
type WhoamiCommand struct{ *cmds.CommandDescription }

type WhoamiSettings struct {
	ConnectionSettings
	Handle string `glazed:"handle"`
}

func NewWhoamiCommand() (*WhoamiCommand, error) {
	desc := cmds.NewCommandDescription(
		"whoami",
		cmds.WithShort("Show your own identity (resolves via --handle or config)"),
		cmds.WithFlags(
			fields.New("handle", fields.TypeString, fields.WithHelp("Your @handle (overrides config)")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &WhoamiCommand{CommandDescription: desc}, nil
}

func (c *WhoamiCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &WhoamiSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	m, err := client.Whoami(ctx, s.Handle)
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, memberRow(m))
}
