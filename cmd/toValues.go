package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

//
// helm dump move-to-values apps/v1 Deployment ".spec.replicas" "{{ .metadata.name }}.replicas"
//
var MoveToValuesCmd = &cobra.Command{
	Use:   "move-to-values",
	Short: "Move a value from a template into values.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("move-to-values called")
	},
}

func init() {
	rootCmd.AddCommand(MoveToValuesCmd)
}
