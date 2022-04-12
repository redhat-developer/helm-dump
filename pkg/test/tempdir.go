package test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TempDir(t *testing.T) string {
	outDir, err := ioutil.TempDir(os.TempDir(), "helm-dump")
	require.NoError(t, err)
	return outDir
}
