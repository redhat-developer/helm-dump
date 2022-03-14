package cmd

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestNewInitCmd(t *testing.T) {
	t.Run("no-arguments", func(t *testing.T) {
		cmd := NewInitCmd()

		cmd.SetArgs([]string{})

		require.Error(t, cmd.Execute(), "Cmd requires an output directory")
	})

	t.Run("minimum-required-arguments", func(t *testing.T) {
		chartName := "my-chart"
		chartVersion := "0.1.0"

		outDir, err := ioutil.TempDir(os.TempDir(), "helm-dump")
		require.NoError(t, err, "temp directory is required for testing")

		cmd := NewInitCmd()

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
