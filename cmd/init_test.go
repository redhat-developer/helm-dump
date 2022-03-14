package cmd

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io/ioutil"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"os"
	"path"
	"testing"
)

func buildConfigFlagsAndRestConfig() (*genericclioptions.ConfigFlags, *rest.Config) {
	configFlags := genericclioptions.NewConfigFlags(true)
	restConfig, err := configFlags.ToRESTConfig()
	if err != nil {
		panic(err)
	}
	return configFlags, restConfig
}

func TestNewInitCmd(t *testing.T) {
	t.Run("no-arguments", func(t *testing.T) {
		configFlags, restConfig := buildConfigFlagsAndRestConfig()
		dynamicClient := dynamic.NewForConfigOrDie(restConfig)
		cmd := NewInitCmd(configFlags, dynamicClient)

		cmd.SetArgs([]string{})

		require.Error(t, cmd.Execute(), "Cmd requires an output directory")
	})

	t.Run("minimum-required-arguments", func(t *testing.T) {
		chartName := "my-chart"
		chartVersion := "0.1.0"

		outDir, err := ioutil.TempDir(os.TempDir(), "helm-dump")
		require.NoError(t, err, "temp directory is required for testing")

		configFlags, restConfig := buildConfigFlagsAndRestConfig()
		dynamicClient := dynamic.NewForConfigOrDie(restConfig)
		cmd := NewInitCmd(configFlags, dynamicClient)

		// helm dump init chart-dir
		cmd.SetArgs([]string{
			"--namespace", "default",
			chartName, outDir})

		require.NoError(t, cmd.Execute(), "Cmd must not return an error")

		expectedChartPath := path.Join(outDir, fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))
		_, err = os.Stat(expectedChartPath)
		require.NoError(t, err, "%q should exist", expectedChartPath)

		_, err = loader.LoadFile(expectedChartPath)
		require.NoError(t, err, "%q should be a chart", expectedChartPath)
	})
}
