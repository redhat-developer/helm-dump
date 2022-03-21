package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kdiscovery "k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"os"
	"path"
	"testing"

	"github.com/redhat-developer/helm-dump/pkg/test"
)

func TestNewInitCmd(t *testing.T) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	t.Run("no-arguments", func(t *testing.T) {
		// Arrange
		configFlags := genericclioptions.NewConfigFlags(true)
		cmd, _ := NewInitCmd(configFlags, logger)
		cmd.SetArgs([]string{})

		// Act & Assert
		require.Error(t, cmd.Execute(), "Cmd requires arguments")
	})

	t.Run("minimum-required-arguments", func(t *testing.T) {
		// Arrange
		chartName := "my-chart"
		chartVersion := "0.1.0"

		outDir, err := ioutil.TempDir(os.TempDir(), "helm-dump")
		require.NoError(t, err, "temp directory is required for testing")

		deployment := test.LoadYamlFixture(t, "init_test/minimum-required-arguments/nginx-deployment.yaml")
		scheme := runtime.NewScheme()

		dynamicClient := fakedynamic.NewSimpleDynamicClient(scheme, deployment)
		discoveryClient := &FakeCachedDiscovery{
			FakeDiscovery: &fakediscovery.FakeDiscovery{
				Fake: &dynamicClient.Fake,
			},
		}
		discoveryClient.Resources = []*metav1.APIResourceList{
			{
				GroupVersion: appsv1.SchemeGroupVersion.String(),
				APIResources: []metav1.APIResource{
					{Name: "deployments", Namespaced: true, Kind: "Deployment", Group: "apps", Version: "v1", Verbs: []string{"list", "create", "get", "delete"}},
				},
			},
		}
		configFlags := genericclioptions.NewConfigFlags(true)

		cmd, _ := NewInitCmd(configFlags, logger)
		cmd.PluginDir = "../plugins/helm_dump_init/dist/"
		cmd.DiscoveryClient = discoveryClient
		cmd.DynamicClient = dynamicClient
		cmd.SetArgs([]string{
			"--namespace", "default",
			chartName, outDir})

		// Act
		require.NoError(t, cmd.Execute(), "Cmd must not return an error")

		// Assert
		expectedChartPath := path.Join(outDir, fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))
		_, err = os.Stat(expectedChartPath)
		require.NoError(t, err, "%q should exist", expectedChartPath)

		chrt, err := loader.LoadFile(expectedChartPath)
		require.NoError(t, err, "%q should be a chart", expectedChartPath)

		require.Len(t, chrt.Templates, 1, "chart should have only one template")
	})
}

type FakeCachedDiscovery struct {
	*fakediscovery.FakeDiscovery
}

func (f *FakeCachedDiscovery) Fresh() bool {
	return true
}

func (f *FakeCachedDiscovery) Invalidate() {
	// no-op
}

var _ kdiscovery.CachedDiscoveryInterface = &FakeCachedDiscovery{}
