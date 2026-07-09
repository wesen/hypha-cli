package hyphamod

import (
	"context"
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/hypha-cli/pkg/hypha"
)

// EventBuilder is the fluent builder for posting an event. Mutators return
// the same *EventBuilder for chaining and append to errs on misuse; Post
// validates first and throws an aggregated error, then calls the client.
type EventBuilder struct {
	rt     *runtime
	opts   hypha.PostOptions
	errs   []string
}

func (r *runtime) event(body string) *EventBuilder {
	return &EventBuilder{rt: r, opts: hypha.PostOptions{Body: body, Audience: "circle"}}
}

func (b *EventBuilder) Kind(kind string) *EventBuilder {
	b.errs = append(b.errs, validateKind(kind)...)
	b.opts.Kind = kind
	return b
}

func (b *EventBuilder) Topic(topics ...string) *EventBuilder {
	b.opts.Topics = append(b.opts.Topics, topics...)
	return b
}

func (b *EventBuilder) To(target string) *EventBuilder { b.opts.Target = target; return b }
func (b *EventBuilder) Ref(ref string) *EventBuilder   { b.opts.Ref = ref; return b }
func (b *EventBuilder) Audience(a string) *EventBuilder { b.opts.Audience = a; return b }
func (b *EventBuilder) IdemKey(k string) *EventBuilder  { b.opts.IdemKey = k; return b }

func (b *EventBuilder) Value(amount float64, unit string) *EventBuilder {
	if unit == "" {
		b.errs = append(b.errs, "value: unit is required (e.g. .value(2, \"hours\"))")
	}
	if unit == "hours" && b.opts.Audience != "circle" {
		b.errs = append(b.errs, fmt.Sprintf("value: unit \"hours\" creates time debt; audience should be \"circle\" (got %q)", b.opts.Audience))
	}
	if b.opts.Target == "" {
		b.errs = append(b.errs, "value: a valued event requires .to(<member>)")
	}
	b.opts.Value = &hypha.Value{Amount: amount, Unit: unit}
	return b
}

func (b *EventBuilder) Validate() *ValidationResult {
	errs := append([]string(nil), b.errs...)
	if strings.TrimSpace(b.opts.Body) == "" && b.opts.Kind != "kudos" {
		errs = append(errs, "event body must not be empty")
	}
	return &ValidationResult{Valid: len(errs) == 0, Errors: errs}
}

// ToSpec returns a plain map for JSON.stringify (builders are not otherwise
// serializable). Mirrors markdown's render-result pattern.
func (b *EventBuilder) ToSpec() map[string]any {
	return map[string]any{
		"body":     b.opts.Body,
		"kind":     b.opts.Kind,
		"topics":   b.opts.Topics,
		"target":   b.opts.Target,
		"ref":      b.opts.Ref,
		"audience": b.opts.Audience,
		"value":    b.opts.Value,
		"idemKey":  b.opts.IdemKey,
	}
}

// Post is the terminal: validate, throw aggregated error, then POST.
func (b *EventBuilder) Post() (*hypha.Event, error) {
	v := b.Validate()
	if !v.Valid {
		return nil, fmt.Errorf("hypha.event: %s", strings.Join(v.Errors, "; "))
	}
	ev, err := b.rt.client.Post(context.Background(), b.opts)
	if err != nil {
		return nil, fmt.Errorf("hypha.event.post: %w", err)
	}
	return ev, nil
}

// throw is a convenience to panic a JS TypeError from a Go error.
func (r *runtime) throw(err error) {
	panic(r.vm.NewTypeError(err.Error()))
}

// objectToMap converts a JS object value to a map[string]any via JSON round-trip.
func (r *runtime) objectToMap(v goja.Value) map[string]any {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return nil
	}
	obj := v.ToObject(r.vm)
	if obj == nil {
		return nil
	}
	out := map[string]any{}
	for _, key := range obj.Keys() {
		out[key] = obj.Get(key).Export()
	}
	return out
}
