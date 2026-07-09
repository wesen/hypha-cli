// Package hyphaprovider is the xgoja provider that selects the "hypha"
// go-go-goja module into generated xgoja binaries. It is the bridge between
// the declarative xgoja.yaml and the native module in pkg/gojamodules/hypha.
package hyphaprovider

import (
	"encoding/json"
	"fmt"

	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/modules"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
	hyphamod "github.com/go-go-golems/hypha-cli/pkg/gojamodules/hypha"
	"github.com/go-go-golems/hypha-cli/pkg/xgoja/providers/hypha/doc"
)

// PackageID is the provider package id used in xgoja.yaml
// (providers[].id: hypha).
const PackageID = "hypha"

// Register exposes the hypha module to xgoja-generated binaries.
// It is referenced from xgoja.yaml as `register: Register`.
func Register(registry *providerapi.ProviderRegistry) error {
	mod := modules.GetModule("hypha")
	if mod == nil {
		return fmt.Errorf("hypha provider: module %q is not registered (blank-import pkg/gojamodules/hypha)", "hypha")
	}
	return registry.Package(PackageID,
		providerapi.Module{
			Name:        "hypha",
			DefaultAs:   "hypha",
			Description: mod.Doc(),
			ConfigSchema: json.RawMessage(`{
  "type": "object",
  "properties": {
    "baseUrl": {"type": "string", "description": "Hypha base URL (e.g. https://hyphahypha.club)"},
    "pat":     {"type": "string", "description": "Personal Access Token (hh_pat_...)"},
    "handle":  {"type": "string", "description": "Your @handle, for whoami"}
  }
}`),
			NewModuleFactory: func(ctx providerapi.ModuleSetupContext) (require.ModuleLoader, error) {
				// Hand the per-module config to the adapter before Loader runs.
				hyphamod.SetModuleConfig(ctx.Config)
				return mod.Loader, nil
			},
		},
		providerapi.HelpSource{
			Name:        "hypha-runtime-api",
			Description: "Hypha JS module API reference (fluent builders + views)",
			FS:          doc.FS(),
			Root:        ".",
		},
	)
}
