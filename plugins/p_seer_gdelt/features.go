package p_seer_gdelt

import (
	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/registry"
	"github.com/UniquityVentures/lamu/views"
)

var (
	pluginPageEntries       []registry.Pair[string, components.PageInterface]
	pluginPagePatches       []registry.Pair[string, func(components.PageInterface) components.PageInterface]
	pluginViewEntries       []registry.Pair[string, *views.View]
	pluginViewPatches       []registry.Pair[string, func(*views.View) *views.View]
	pluginRouteEntries      []registry.Pair[string, lamu.Route]
	pluginRoutePatches      []registry.Pair[string, func(lamu.Route) lamu.Route]
	pluginConfigEntries     []registry.Pair[string, lamu.Config]
	pluginDBInitHookEntries []registry.Pair[string, lamu.DBInitHook]
)

func registerPluginPage(key string, value components.PageInterface) {
	pluginPageEntries = append(pluginPageEntries, registry.Pair[string, components.PageInterface]{Key: key, Value: value})
}

func patchPluginPage(key string, patch func(components.PageInterface) components.PageInterface) {
	pluginPagePatches = append(pluginPagePatches, registry.Pair[string, func(components.PageInterface) components.PageInterface]{Key: key, Value: patch})
}

func registerPluginView(key string, value *views.View) {
	pluginViewEntries = append(pluginViewEntries, registry.Pair[string, *views.View]{Key: key, Value: value})
}

func patchPluginView(key string, patch func(*views.View) *views.View) {
	pluginViewPatches = append(pluginViewPatches, registry.Pair[string, func(*views.View) *views.View]{Key: key, Value: patch})
}

func registerPluginRoute(key string, value lamu.Route) {
	pluginRouteEntries = append(pluginRouteEntries, registry.Pair[string, lamu.Route]{Key: key, Value: value})
}

func patchPluginRoute(key string, patch func(lamu.Route) lamu.Route) {
	pluginRoutePatches = append(pluginRoutePatches, registry.Pair[string, func(lamu.Route) lamu.Route]{Key: key, Value: patch})
}

func registerPluginConfig(key string, value lamu.Config) {
	pluginConfigEntries = append(pluginConfigEntries, registry.Pair[string, lamu.Config]{Key: key, Value: value})
}

func registerPluginDBInitHook(key string, value lamu.DBInitHook) {
	pluginDBInitHookEntries = append(pluginDBInitHookEntries, registry.Pair[string, lamu.DBInitHook]{Key: key, Value: value})
}

func pluginPages() lamu.PluginFeatures[components.PageInterface] {
	return lamu.PluginFeatures[components.PageInterface]{Entries: pluginPageEntries, Patches: pluginPagePatches}
}

func pluginViews() lamu.PluginFeatures[*views.View] {
	return lamu.PluginFeatures[*views.View]{Entries: pluginViewEntries, Patches: pluginViewPatches}
}

func pluginRoutes() lamu.PluginFeatures[lamu.Route] {
	return lamu.PluginFeatures[lamu.Route]{Entries: pluginRouteEntries, Patches: pluginRoutePatches}
}

func pluginConfigs() lamu.PluginFeatures[lamu.Config] {
	return lamu.PluginFeatures[lamu.Config]{Entries: pluginConfigEntries}
}

func pluginDBInitHooks() lamu.PluginFeatures[lamu.DBInitHook] {
	return lamu.PluginFeatures[lamu.DBInitHook]{Entries: pluginDBInitHookEntries}
}
