package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"helm.sh/helm/v3/pkg/chartutil"
)

type MoveToValuesCommand struct {
	*cobra.Command
	Logger      *logrus.Logger
	ProjectRoot string
	OutputDir   string
}

func NewMoveToValuesCmd(logger *logrus.Logger) (*MoveToValuesCommand, error) {
	cmd := &MoveToValuesCommand{
		Logger: logger,
		Command: &cobra.Command{
			Use:   "move-to-values api-version kind field template",
			Short: "Move a value from a template into values.yaml",
			Args:  cobra.ExactArgs(4),
		},
	}

	cmd.PersistentFlags().StringVarP(&cmd.ProjectRoot, "project-root", "d", ".", "The project root directory")
	cmd.PersistentFlags().StringVarP(&cmd.OutputDir, "output-directory", "o", "", "The output directory; if unspecified overwrites file in project-root")

	cmd.Command.RunE = cmd.runE

	return cmd, nil
}

func (c *MoveToValuesCommand) runE(cmd *cobra.Command, args []string) error {

	chartBuilder, err := NewChartBuilder(c.ProjectRoot, c.OutputDir)
	if err != nil {
		return fmt.Errorf("error creating builder: %w", err)
	}

	chartBuilder.AddAction(&Action{
		apiVersion: args[0],
		kind:       args[1],
		path:       args[2],
		template:   args[3],
	})

	chrt, buildErr := chartBuilder.Build()
	if buildErr != nil {
		return fmt.Errorf("error building chart: %w", buildErr)
	}

	saveErr := chartutil.SaveDir(chrt, c.OutputDir)
	if saveErr != nil {
		return fmt.Errorf("error saving chart: %w", saveErr)
	}

	return nil
}

func init() {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	cmd, err := NewMoveToValuesCmd(logger)
	if err != nil {
		panic(err)
	}
	rootCmd.AddCommand(cmd.Command)
}
