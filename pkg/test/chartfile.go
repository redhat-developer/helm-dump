package test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// RequireChartFileExists ...
func RequireChartFileExists(
	t *testing.T,
	baseDir, chartName, chartVersion string,
) {
	expectedChartPath := path.Join(baseDir, fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))
	_, err := os.Stat(expectedChartPath)
	require.NoError(t, err, "%q should exist", expectedChartPath)
}

// RequireChart ...
func RequireChart(t *testing.T, baseDir, chartName, chartVersion string) *chart.Chart {
	expectedChartPath := path.Join(baseDir, fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))
	chrt, err := loader.LoadFile(expectedChartPath)
	require.NoError(t, err, "%q should be a chart", expectedChartPath)
	return chrt
}
