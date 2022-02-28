package plugin

import (
	"github.com/redhat-developer/helm-dump/pkg/helm/plugin/env"
)

type Settings struct {
	// Kubeconfig is the path of the Kubernetes client configuration to be used by the plugin.
	Kubeconfig string
	// PluginName is the name of the plugin, as invoked by helm. So helm myplug will have the
	// short name myplug.
	PluginName string
	// PluginsDirectory is the path to the plugins directory.
	PluginsDirectory string
	// PluginDirectory is the directory that contains the plugin.
	PluginDirectory string
	// ProgramPath is the path to the helm command (as executed by the user).
	ProgramPath string
	// Debug indicates if the debug flag was set by helm.
	Debug bool
	// RegistryConfig is the location for the registry configuration (if using). Note that the
	// use of Helm with registries is an experimental feature.
	RegistryConfig string
	// RepositoryCache is the path to the repository cache files.
	RepositoryCache string
	// RepositoryConfig is the path to the repository configuration file.
	RepositoryConfig string
	// Namespace is the Namespace given to the helm command (generally using the -n flag).
	Namespace string
	// KubeContext is the name of the Kubernetes config context given to the helm command.
	KubeContext string
	// MaxHistory is the max release history maintained.
	MaxHistory int
}

func NewSettings() *Settings {
	return &Settings{
		Kubeconfig:       env.String("KUBECONFIG"),
		PluginsDirectory: env.String("HELM_PLUGINS"),
		PluginDirectory:  env.String("HELM_PLUGIN_DIR"),
		ProgramPath:      env.String("HELM_BIN"),
		RegistryConfig:   env.String("HELM_REGISTRY_CONFIG"),
		RepositoryCache:  env.String("HELM_REPOSITORY_CACHE"),
		RepositoryConfig: env.String("HELM_REPOSITORY_CONFIG"),
		Namespace:        env.String("HELM_NAMESPACE"),
		KubeContext:      env.String("HELM_KUBECONTEXT"),
		MaxHistory:       env.Int("HELM_MAX_HISTORY"),
	}
}
