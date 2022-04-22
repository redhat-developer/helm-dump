package cmd

import (
	"path/filepath"
	"testing"

	hdtesting "github.com/redhat-developer/helm-dump/pkg/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/apimachinery/pkg/api/equality"
)

func makeMoveToValuesCommandWorld(logger *logrus.Logger) (*MoveToValuesCommand, error) {
	return NewMoveToValuesCmd(logger)
}

func TestMoveToValuesCmd(t *testing.T) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	t.Run("no-arguments", func(t *testing.T) {
		// Arrange
		cmd, err := NewMoveToValuesCmd(logger)
		require.NoError(t, err)
		cmd.SetArgs([]string{})

		// Act & Assert
		require.Error(t, cmd.Execute(), "Cmd requires arguments")
	})

	t.Run("minimum-required-arguments", func(t *testing.T) {
		// Arrange
		tempDir := hdtesting.TempDir(t)
		inputDir := "move_to_values_test/minimum-required-arguments/input-chart"
		expectedDir := "move_to_values_test/minimum-required-arguments/expected-chart"

		cmd, err := NewMoveToValuesCmd(logger)
		require.NoError(t, err)
		cmd.SetArgs([]string{
			"-d", inputDir,
			"-o", tempDir,
			"apps/v1",
			"Deployment",
			`.spec.replicas`,
			`{{ .metadata.name }}.replicas`,
		})

		// Act
		require.NoError(t, cmd.Execute())

		// Assert
		expectedChart, err := loader.LoadDir(expectedDir)
		require.NoError(t, err)

		actualChartDir := filepath.Join(tempDir, expectedChart.Name())

		actualChart, err := loader.LoadDir(actualChartDir)
		require.NoError(t, err, "chart should exist in %q", actualChartDir)

		if !equality.Semantic.DeepEqual(expectedChart, actualChart) {
			require.Equal(t,
				string(expectedChart.Templates[1].Data),
				string(actualChart.Templates[1].Data),
			)

			diff, err := hdtesting.YamlDiff(expectedChart, actualChart)
			require.NoError(t, err)
			t.Errorf("expected different than actual:\n%s", diff)
		}
	})
}
