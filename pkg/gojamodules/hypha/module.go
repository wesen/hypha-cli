// Package hyphamod is the go-go-goja native module adapter that exposes the
// Hypha client to JavaScript as a fluent builder API backed by Go-side opaque
// objects (the goja-text markdown style). require("hypha") returns a small
// set of factory functions whose results are Go pointers; goja reflects
// their exported methods/fields into JS for chaining.
package hyphamod

import (
	"encoding/json"
	"os"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/modules"
	"github.com/go-go-golems/hypha-cli/pkg/hypha"
)

type module struct{}

var _ modules.NativeModule = (*module)(nil)

func (module) Name() string { return "hypha" }

func (module) Doc() string {
	return `Hypha kernel fluent client.

require("hypha") returns factory functions that build Go-backed objects:
  hypha.event("body").kind("update").topic("meta").post()   -> Event
  hypha.feed().kind("update").limit(20).list()              -> Feed
  hypha.views.balance(); hypha.whois("@alice"); hypha.pats().scopes("read").mint()

Builders accumulate validation errors and throw an aggregated error at the
terminal call (Post/List/Mint/Run). Result objects expose Go fields: event.ID,
member.Handle, checkpoint.Hash, ...`
}

// moduleConfig is the per-module config from xgoja.yaml runtime.modules[].config.
type moduleConfig struct {
	BaseURL string `json:"baseUrl"`
	PAT     string `json:"pat"`
	Handle  string `json:"handle"`
}

// moduleConfigRaw is set by the xgoja provider before Loader runs.
var moduleConfigRaw json.RawMessage

// SetModuleConfig is called by the xgoja provider to hand per-module config
// to the adapter before the runtime loads the module.
func SetModuleConfig(raw json.RawMessage) { moduleConfigRaw = raw }

func (module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	cfg := moduleConfig{
		BaseURL: os.Getenv("HYPHA_BASE_URL"),
		PAT:     os.Getenv("HYPHA_PAT"),
		Handle:  os.Getenv("HYPHA_HANDLE"),
	}
	if len(moduleConfigRaw) > 0 {
		_ = json.Unmarshal(moduleConfigRaw, &cfg)
	}
	client := hypha.NewClient(cfg.BaseURL, cfg.PAT)
	rt := &runtime{vm: vm, client: client, handle: cfg.Handle}

	exports := moduleObj.Get("exports").(*goja.Object)
	rt.mustSet(exports, "event", rt.event)
	rt.mustSet(exports, "feed", rt.feed)
	rt.mustSet(exports, "search", rt.search)
	rt.mustSet(exports, "whois", rt.whois)
	rt.mustSet(exports, "whoami", rt.whoami)
	rt.mustSet(exports, "members", rt.members)
	rt.mustSet(exports, "kudos", rt.kudos)
	rt.mustSet(exports, "webhook", rt.webhook)
	rt.mustSet(exports, "webhooks", rt.webhooksList)
	rt.mustSet(exports, "unsubscribe", rt.unsubscribe)
	rt.mustSet(exports, "pats", rt.pats)
	rt.mustSet(exports, "export", rt.exportQuery)
	rt.mustSet(exports, "checkpoints", rt.checkpointsQuery)
	rt.mustSet(exports, "verify", rt.verifyQuery)

	// Asks (ISO board) via MCP /mcp tools.
	rt.mustSet(exports, "postISO", rt.postISO)
	rt.mustSet(exports, "replyAsk", rt.replyAsk)
	rt.mustSet(exports, "closeAsk", rt.closeAsk)
	rt.mustSet(exports, "listAsks", rt.listAsks)
	rt.mustSet(exports, "getAsk", rt.getAsk)

	views := vm.NewObject()
	rt.mustSet(views, "balance", rt.balance)
	rt.mustSet(views, "trust", rt.trust)
	rt.mustSet(views, "topics", rt.topics)
	rt.mustSet(views, "whoCanHelp", rt.whoCanHelp)
	rt.mustSet(views, "gratitude", rt.gratitude)
	rt.mustSet(exports, "views", views)
}

func init() { modules.Register(module{}) }

// runtime holds the shared client and JS runtime for one module load.
type runtime struct {
	vm     *goja.Runtime
	client *hypha.Client
	handle string
}

func (r *runtime) mustSet(o *goja.Object, key string, value any) {
	if err := o.Set(key, value); err != nil {
		panic(r.vm.NewGoError(err))
	}
}
