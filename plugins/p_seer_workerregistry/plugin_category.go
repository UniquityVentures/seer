package p_seer_workerregistry

import (
	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/lago/registry"
)

// RegistryPluginCategory groups Seer dashboard apps by category key (e.g. "Sources", "Live Maps").
// Each value is the ordered list of plugins in that category; use [RegisterPluginInCategory] from init().
var RegistryPluginCategory *registry.Registry[[]lago.Plugin] = registry.NewRegistry[[]lago.Plugin]()

// RegisterPluginInCategory appends p to the slice for category, creating the category on first use.
func RegisterPluginInCategory(category string, p lago.Plugin) {
	if err := RegistryPluginCategory.Register(category, []lago.Plugin{p}); err != nil {
		RegistryPluginCategory.Patch(category, func(s []lago.Plugin) []lago.Plugin {
			return append(s, p)
		})
	}
}

// BuildAllRegistries materializes Seer-side registries after all plugin init() functions have run.
// Call this from the Seer main package before [lago.Start] (see deployments/seer main.go).
func BuildAllRegistries() {
	_ = RegistryActiveWorkersProvider.All()
	_ = RegistryPluginCategory.All()
}
