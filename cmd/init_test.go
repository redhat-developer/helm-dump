package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"os"
	"path"
	"testing"
)

func buildClients() (*fakedynamic.FakeDynamicClient, *fakediscovery.FakeDiscovery) {
	scheme := runtime.NewScheme()
	dynamicClient := fakedynamic.NewSimpleDynamicClient(scheme)
	discoveryClient := &fakediscovery.FakeDiscovery{
		Fake: &dynamicClient.Fake,
	}

	return dynamicClient, discoveryClient
}

func TestNewInitCmd(t *testing.T) {
	logger, _ := test.NewNullLogger()

	t.Run("no-arguments", func(t *testing.T) {
		// Arrange
		dynamicClient, discoveryClient := buildClients()

		configFlags := genericclioptions.NewConfigFlags(true)
		cmd, _ := NewInitCmd(configFlags, dynamicClient, discoveryClient, logger)
		cmd.SetArgs([]string{})

		// Act & Assert
		require.Error(t, cmd.Execute(), "Cmd requires an output directory")
	})

	t.Run("minimum-required-arguments", func(t *testing.T) {
		// Arrange
		chartName := "my-chart"
		chartVersion := "0.1.0"

		outDir, err := ioutil.TempDir(os.TempDir(), "helm-dump")
		require.NoError(t, err, "temp directory is required for testing")

		dynamicClient, discoveryClient := buildClients()
		configFlags := genericclioptions.NewConfigFlags(true)
		cmd, _ := NewInitCmd(configFlags, dynamicClient, discoveryClient, logger)
		cmd.SetArgs([]string{
			"--namespace", "default",
			chartName, outDir})

		// Act
		require.NoError(t, cmd.Execute(), "Cmd must not return an error")

		// Assert
		expectedChartPath := path.Join(outDir, fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))
		_, err = os.Stat(expectedChartPath)
		require.NoError(t, err, "%q should exist", expectedChartPath)

		_, err = loader.LoadFile(expectedChartPath)
		require.NoError(t, err, "%q should be a chart", expectedChartPath)
	})
}
