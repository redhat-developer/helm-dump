package cmd

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	t.Run("without-arguments", func(t *testing.T) {
		expected := &BuildInfo{Version: "0.1.0"}
		cmd := NewVersionCmd(expected)

		// args is set to os.Args[1:] by default, not useful here.
		cmd.SetArgs([]string{})

		// outBuf captures the command's Stdout.
		outBuf := bytes.NewBufferString("")
		cmd.SetOut(outBuf)

		// errBuf captures the command's Stderr.
		errBuf := bytes.NewBufferString("")
		cmd.SetErr(errBuf)

		require.NoError(t, cmd.Execute(), "Cmd must not return an error")

		actual := &BuildInfo{}
		err := json.Unmarshal(outBuf.Bytes(), &actual)
		require.NoError(t, err, "Cmd output must fit in BuildInfo")

		require.Equal(t, expected, actual, "Cmd output content must match the cmd's input value")
	})
}
