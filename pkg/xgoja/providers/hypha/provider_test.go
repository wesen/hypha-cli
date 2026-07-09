package hyphaprovider_test

import (
	"context"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/modules"
	gggengine "github.com/go-go-golems/go-go-goja/pkg/engine"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
	_ "github.com/go-go-golems/hypha-cli/pkg/gojamodules/hypha" // registers the module
	hyphaprovider "github.com/go-go-golems/hypha-cli/pkg/xgoja/providers/hypha"
	"github.com/stretchr/testify/require"
)

func TestProviderRegisterAndResolve(t *testing.T) {
	reg := providerapi.NewProviderRegistry()
	require.NoError(t, hyphaprovider.Register(reg))
	mod, ok := reg.ResolveModule(hyphaprovider.PackageID, "hypha")
	require.True(t, ok, "hypha module should resolve")
	require.Equal(t, "hypha", mod.Name)
	require.NotNil(t, mod.NewModuleFactory)
}

func TestProviderModuleFactoryLoads(t *testing.T) {
	reg := providerapi.NewProviderRegistry()
	require.NoError(t, hyphaprovider.Register(reg))
	mod, ok := reg.ResolveModule(hyphaprovider.PackageID, "hypha")
	require.True(t, ok)
	loader, err := mod.NewModuleFactory(providerapi.ModuleSetupContext{
		Context: context.Background(),
		Name:    "hypha",
		As:      "hypha",
	})
	require.NoError(t, err)
	require.NotNil(t, loader)

	// Boot a runtime and confirm require("hypha") resolves via the loader.
	factory, err := gggengine.NewRuntimeFactoryBuilder().Build()
	require.NoError(t, err)
	rt, err := factory.NewRuntime(
		gggengine.WithStartupContext(context.Background()),
		gggengine.WithLifetimeContext(context.Background()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = rt.Close(context.Background()) })

	_, err = rt.Owner.Call(context.Background(), "test", func(_ context.Context, vm *goja.Runtime) (any, error) {
		v, err := vm.RunString(`typeof require("hypha")`)
		if err != nil {
			return nil, err
		}
		return v.Export(), nil
	})
	require.NoError(t, err)
}

func TestModuleIsRegistered(t *testing.T) {
	// The blank import above should have registered the native module.
	require.NotNil(t, modules.GetModule("hypha"))
}
