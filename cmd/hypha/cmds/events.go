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

// eventRow turns a Hypha Event into a Glazed row. Pointer fields are
// dereferenced to their value (or nil when absent) so formatters don't print
// pointer addresses.
func eventRow(ev *hypha.Event) types.Row {
	return types.NewRow(
		types.MRP("id", ev.ID),
		types.MRP("ts", ev.Ts),
		types.MRP("actor", ev.Actor),
		types.MRP("verb", ev.Verb),
		types.MRP("kind", derefStr(ev.Kind)),
		types.MRP("target", derefStr(ev.Target)),
		types.MRP("ref", derefStr(ev.Ref)),
		types.MRP("audience", ev.Audience),
		types.MRP("topics", ev.Topics),
		types.MRP("body", derefStr(ev.Body)),
		types.MRP("value_amount", derefFloat(ev.ValueAmount)),
		types.MRP("value_unit", derefStr(ev.ValueUnit)),
		types.MRP("redacted", ev.Redacted),
	)
}

func derefStr(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

func derefFloat(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

// PostCommand implements `hypha post <body> [flags]`.
type PostCommand struct{ *cmds.CommandDescription }

type PostSettings struct {
	ConnectionSettings
	Body     string   `glazed:"body"`
	Kind     string   `glazed:"kind"`
	Topic    []string `glazed:"topic"`
	To       string   `glazed:"to"`
	Ref      string   `glazed:"ref"`
	Audience string   `glazed:"audience"`
	Value    float64  `glazed:"value"`
	Unit     string   `glazed:"unit"`
	IdemKey  string   `glazed:"idem-key"`
}

func NewPostCommand() (*PostCommand, error) {
	desc := cmds.NewCommandDescription(
		"post",
		cmds.WithShort("Append an event to the circle log"),
		cmds.WithLong(`Append an event to the append-only log (POST /api/v1/events).

Examples:
  hypha post "hello circle" --topic meta --kind update
  hypha post "gave alice 2h" --to @alice --value 2 --unit hours --kind gig`),
		cmds.WithArguments(
			fields.New("body", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("Event body text"), fields.WithIsArgument(true)),
		),
		cmds.WithFlags(
			fields.New("kind", fields.TypeString, fields.WithHelp("Open kind tag; dot-prefix third-party (e.g. acme.bounty)")),
			fields.New("topic", fields.TypeStringList, fields.WithHelp("Topic tag (repeatable)")),
			fields.New("to", fields.TypeString, fields.WithHelp("Target member id or @handle")),
			fields.New("ref", fields.TypeString, fields.WithHelp("Referenced event id")),
			fields.New("audience", fields.TypeString, fields.WithDefault("circle"), fields.WithHelp("circle or a member id")),
			fields.New("value", fields.TypeFloat, fields.WithHelp("Value amount")),
			fields.New("unit", fields.TypeString, fields.WithHelp("hours | kudos | <namespaced>")),
			fields.New("idem-key", fields.TypeString, fields.WithHelp("Idempotency key")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &PostCommand{CommandDescription: desc}, nil
}

func (c *PostCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &PostSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	opts := hypha.PostOptions{
		Body:     s.Body,
		Kind:     s.Kind,
		Topics:   s.Topic,
		Target:   s.To,
		Ref:      s.Ref,
		Audience: s.Audience,
		IdemKey:  s.IdemKey,
	}
	if s.Unit != "" {
		opts.Value = &hypha.Value{Amount: s.Value, Unit: s.Unit}
	}
	ev, err := client.Post(ctx, opts)
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, eventRow(ev))
}

// FeedCommand implements `hypha feed [flags]`.
type FeedCommand struct{ *cmds.CommandDescription }

type FeedSettings struct {
	ConnectionSettings
	Kind   string   `glazed:"kind"`
	Topic  []string `glazed:"topic"`
	Actor  string   `glazed:"actor"`
	Ref    string   `glazed:"ref"`
	Since  string   `glazed:"since"`
	Limit  int      `glazed:"limit"`
}

func NewFeedCommand() (*FeedCommand, error) {
	desc := cmds.NewCommandDescription(
		"feed",
		cmds.WithShort("Read the event feed"),
		cmds.WithLong(`Read the event feed with filters (GET /api/v1/events).`),
		cmds.WithFlags(
			fields.New("kind", fields.TypeString, fields.WithHelp("Filter by kind")),
			fields.New("topic", fields.TypeStringList, fields.WithHelp("Filter by topic (repeatable)")),
			fields.New("actor", fields.TypeString, fields.WithHelp("Filter by actor id or @handle")),
			fields.New("ref", fields.TypeString, fields.WithHelp("Filter by referenced event id")),
			fields.New("since", fields.TypeString, fields.WithHelp("Filter since (event id or timestamp)")),
			fields.New("limit", fields.TypeInteger, fields.WithHelp("Max results")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &FeedCommand{CommandDescription: desc}, nil
}

func (c *FeedCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &FeedSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	events, err := client.Feed(ctx, hypha.FeedOptions{Kind: s.Kind, Topics: s.Topic, Actor: s.Actor, Ref: s.Ref, Since: s.Since, Limit: s.Limit})
	if err != nil {
		return err
	}
	for i := range events {
		if err := gp.AddRow(ctx, eventRow(&events[i])); err != nil {
			return err
		}
	}
	return nil
}

// SearchCommand implements `hypha search <query>`.
type SearchCommand struct{ *cmds.CommandDescription }

type SearchSettings struct {
	ConnectionSettings
	Query string `glazed:"query"`
	Limit int    `glazed:"limit"`
}

func NewSearchCommand() (*SearchCommand, error) {
	desc := cmds.NewCommandDescription(
		"search",
		cmds.WithShort("Full-text search over the feed"),
		cmds.WithArguments(
			fields.New("query", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("Search query"), fields.WithIsArgument(true)),
		),
		cmds.WithFlags(
			fields.New("limit", fields.TypeInteger, fields.WithHelp("Max results")),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &SearchCommand{CommandDescription: desc}, nil
}

func (c *SearchCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &SearchSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	hits, err := client.Search(ctx, s.Query, s.Limit)
	if err != nil {
		return err
	}
	for i := range hits {
		if err := gp.AddRow(ctx, eventRow(&hits[i])); err != nil {
			return err
		}
	}
	return nil
}

// RedactCommand implements `hypha redact <event-id>`.
type RedactCommand struct{ *cmds.CommandDescription }

type RedactSettings struct {
	ConnectionSettings
	ID string `glazed:"id"`
}

func NewRedactCommand() (*RedactCommand, error) {
	desc := cmds.NewCommandDescription(
		"redact",
		cmds.WithShort("Redact an event (blank body/topics, keep the fact)"),
		cmds.WithArguments(
			fields.New("id", fields.TypeString, fields.WithRequired(true),
				fields.WithHelp("Event id to redact"), fields.WithIsArgument(true)),
		),
	)
	if err := AddConnectionSections(desc); err != nil {
		return nil, err
	}
	return &RedactCommand{CommandDescription: desc}, nil
}

func (c *RedactCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &RedactSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if err := vals.DecodeSectionInto(ConnectionSectionName, s); err != nil {
		return err
	}
	client := hypha.NewClient(s.BaseURL, s.PAT)
	if err := client.Redact(ctx, s.ID); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("id", s.ID), types.MRP("redacted", true)))
}
