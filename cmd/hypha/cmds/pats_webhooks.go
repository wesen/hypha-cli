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

// --- PATs (hypha pats <mint|list|revoke>) ---

// PatsMintCommand implements `hypha pats mint --scopes ... --label ...`.
type PatsMintCommand struct{ *cmds.CommandDescription }

type patsMintSettings struct {
	ConnectionSettings
	Scopes []string `glazed:"scopes"`
	Label  string   `glazed:"label"`
}

func NewPatsMintCommand() (*PatsMintCommand, error) {
	desc := cmds.NewCommandDescription(
		"mint",
		cmds.WithShort("Mint a narrower PAT (scopes must subset yours)"),
		cmds.WithFlags(
			fields.New("scopes", fields.TypeStringList, fields.WithRequired(true), fields.WithHelp("Scopes (read|write|value|graph)")),
			fields.New("label", fields.TypeString, fields.WithHelp("Human-readable label")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &PatsMintCommand{CommandDescription: desc}, nil
}

func (c *PatsMintCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &patsMintSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	res, err := client.MintPAT(ctx, hypha.MintPATOptions{Scopes: s.Scopes, Label: s.Label})
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("token", res.Token), types.MRP("id", res.ID)))
}

// PatsListCommand implements `hypha pats list`.
type PatsListCommand struct{ *cmds.CommandDescription }

func NewPatsListCommand() (*PatsListCommand, error) {
	desc := cmds.NewCommandDescription("list", cmds.WithShort("List your PATs (tokens never re-shown)"))
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &PatsListCommand{CommandDescription: desc}, nil
}

func (c *PatsListCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &ConnectionSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	pats, err := client.ListPATs(ctx)
	if err != nil {
		return err
	}
	for i := range pats {
		p := pats[i]
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("id", p.ID), types.MRP("scopes", p.Scopes),
			types.MRP("label", p.Label), types.MRP("created_at", p.CreatedAt),
			types.MRP("revoked_at", p.RevokedAt),
		)); err != nil {
			return err
		}
	}
	return nil
}

// PatsRevokeCommand implements `hypha pats revoke <id>`.
type PatsRevokeCommand struct{ *cmds.CommandDescription }

type patsRevokeSettings struct {
	ConnectionSettings
	ID string `glazed:"id"`
}

func NewPatsRevokeCommand() (*PatsRevokeCommand, error) {
	desc := cmds.NewCommandDescription(
		"revoke",
		cmds.WithShort("Revoke a PAT immediately"),
		cmds.WithArguments(
			fields.New("id", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("PAT id to revoke"), fields.WithIsArgument(true)),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &PatsRevokeCommand{CommandDescription: desc}, nil
}

func (c *PatsRevokeCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &patsRevokeSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	if err := client.RevokePAT(ctx, s.ID); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("id", s.ID), types.MRP("revoked", true)))
}

// --- Webhooks (hypha webhooks <subscribe|list|unsubscribe>) ---

// WebhookSubscribeCommand implements `hypha webhooks subscribe <url> [--secret]`.
type WebhookSubscribeCommand struct{ *cmds.CommandDescription }

type webhookSubscribeSettings struct {
	ConnectionSettings
	URL    string `glazed:"url"`
	Secret string `glazed:"secret"`
}

func NewWebhookSubscribeCommand() (*WebhookSubscribeCommand, error) {
	desc := cmds.NewCommandDescription(
		"subscribe",
		cmds.WithShort("Register a webhook (HMAC-signed, SSRF-guarded)"),
		cmds.WithArguments(
			fields.New("url", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("Webhook URL"), fields.WithIsArgument(true)),
		),
		cmds.WithFlags(
			fields.New("secret", fields.TypeString, fields.WithHelp("HMAC secret")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &WebhookSubscribeCommand{CommandDescription: desc}, nil
}

func (c *WebhookSubscribeCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &webhookSubscribeSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	wh, err := client.Subscribe(ctx, hypha.SubscribeOptions{URL: s.URL, Secret: s.Secret})
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("id", wh.ID), types.MRP("url", wh.URL), types.MRP("created_at", wh.CreatedAt)))
}

// WebhooksListCommand implements `hypha webhooks list`.
type WebhooksListCommand struct{ *cmds.CommandDescription }

func NewWebhooksListCommand() (*WebhooksListCommand, error) {
	desc := cmds.NewCommandDescription("list", cmds.WithShort("List your webhooks"))
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &WebhooksListCommand{CommandDescription: desc}, nil
}

func (c *WebhooksListCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &ConnectionSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	whs, err := client.Webhooks(ctx)
	if err != nil {
		return err
	}
	for i := range whs {
		wh := whs[i]
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("id", wh.ID), types.MRP("url", wh.URL), types.MRP("created_at", wh.CreatedAt),
		)); err != nil {
			return err
		}
	}
	return nil
}

// WebhookUnsubscribeCommand implements `hypha webhooks unsubscribe <id>`.
type WebhookUnsubscribeCommand struct{ *cmds.CommandDescription }

type webhookUnsubscribeSettings struct {
	ConnectionSettings
	ID string `glazed:"id"`
}

func NewWebhookUnsubscribeCommand() (*WebhookUnsubscribeCommand, error) {
	desc := cmds.NewCommandDescription(
		"unsubscribe",
		cmds.WithShort("Delete a webhook"),
		cmds.WithArguments(
			fields.New("id", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("Webhook id"), fields.WithIsArgument(true)),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &WebhookUnsubscribeCommand{CommandDescription: desc}, nil
}

func (c *WebhookUnsubscribeCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &webhookUnsubscribeSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	if err := client.Unsubscribe(ctx, s.ID); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("id", s.ID), types.MRP("unsubscribed", true)))
}
