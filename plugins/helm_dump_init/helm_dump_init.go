package main

import (
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/konveyor/crane-lib/transform"
	"github.com/konveyor/crane-lib/transform/cli"
	"strings"
)

func main() {
	var fields []transform.OptionalFields
	cli.RunAndExit(cli.NewCustomPlugin("HelmDumpInit", "v1", fields, Run))
}

var helmNameFmt = "%s-{{ .Release.Name }}"

func Run(request transform.PluginRequest) (transform.PluginResponse, error) {
	obj := request.Unstructured

	var opsJSON []string

	// patch the object's name accordingly
	newName := fmt.Sprintf(helmNameFmt, obj.GetName())
	opsJSON = append(opsJSON, fmt.Sprintf(`{"op": "replace", "path": "/metadata/name", "value": %q}`, newName))

	if labels := obj.GetLabels(); labels == nil {
		opsJSON = append(opsJSON, `{"op": "add", "path": "/metadata/labels", "value": {}}`)
	}

	opsJSON = append(opsJSON,
		`{"op": "replace", "path": "/metadata/labels/app.kubernetes.io~1name", "value": "{{ template \"fullname\" $ }}"}`,
		`{"op": "replace", "path": "/metadata/labels/app.kubernetes.io~1instance", "value": "{{ $.Release.Name }}"}`,
	)

	patchJSON := fmt.Sprintf("[%s]", strings.Join(opsJSON, ","))

	patch, err := jsonpatch.DecodePatch([]byte(patchJSON))
	if err != nil {
		return transform.PluginResponse{}, err
	}
	resp := transform.PluginResponse{
		Version:    "v1",
		IsWhiteOut: false,
		Patches:    patch,
	}
	return resp, nil
}
