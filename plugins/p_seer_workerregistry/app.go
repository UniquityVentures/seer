package p_seer_workerregistry

import (
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/registry"
)

func GetPlugin() registry.Pair[string, lamu.Plugin] {
	return registry.Pair[string, lamu.Plugin]{
		Key: "p_seer_workerregistry",
		Value: lamu.Plugin{
			DBInitHooks: lamu.PluginStages(pluginDBInitHooks),
		},
	}
}
