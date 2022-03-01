package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	version   string
	commit    string
	date      string
	goVersion string
)

type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
}

func init() {
	info := &BuildInfo{
		Version:   version,
		Commit:    commit,
		Date:      date,
		GoVersion: goVersion,
	}
	rootCmd.AddCommand(NewVersionCmd(info))
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
