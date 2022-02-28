package cmd

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
)

//go:embed build.json
var embeddedBuildInfo []byte

type BuildInfo struct {
	Version string `json:"version"`
}

func init() {
	var info BuildInfo
	err := json.Unmarshal(embeddedBuildInfo, &info)
	if err != nil {
		panic(fmt.Errorf("couldn't find build information: %w", err))
	}
	rootCmd.AddCommand(NewVersionCmd(&info))
}

func NewVersionCmd(info *BuildInfo) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "print the helm-dump plugin version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var out bytes.Buffer
			// errors are ignored here since the process would have panicked earlier in init
			// if buildInfo is not available.
			b, _ := json.Marshal(info)
			_ = json.Indent(&out, b, "", "  ")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", out.Bytes())
		},
	}
	return cmd
}
