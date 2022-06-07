package main

import (
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/konveyor/crane-lib/transform"
	"github.com/konveyor/crane-lib/transform/cli"
	"strings"
)

func main() {
	fields := []transform.OptionalFields{
		{
			FlagName: "chart-name",
		},
	}
	cli.RunAndExit(cli.NewCustomPlugin("HelmDumpInit", "v1", fields, Run))
}

var helmNameFmt = "%s-{{ .Release.Name }}"

func Run(request transform.PluginRequest) (transform.PluginResponse, error) {
	obj := request.Unstructured
	resp := transform.PluginResponse{
		Version: "v1",
		Patches: nil,
	}

	chartName, ok := request.Extras["chart-name"]
	if !ok {
		return transform.PluginResponse{}, fmt.Errorf("chart-name should be informed")
	}

	var opsJSON []string

	// patch the object's name accordingly
	newName := fmt.Sprintf(helmNameFmt, obj.GetName())
	opsJSON = append(opsJSON, fmt.Sprintf(`{"op": "add", "path": "/metadata/name", "value": %q}`, newName))

	if labels := obj.GetLabels(); labels == nil {
		opsJSON = append(opsJSON, `{"op": "add", "path": "/metadata/labels", "value": {}}`)
	}

	opsJSON = append(opsJSON,
		fmt.Sprintf(`{"op": "add", "path": "/metadata/labels/app.kubernetes.io~1name", "value": "{{ template \"%s.fullname\" $ }}"}`, chartName),
		`{"op": "add", "path": "/metadata/labels/app.kubernetes.io~1instance", "value": "{{ $.Release.Name }}"}`,
	)

	// keep the original name for modifications in other commands (so removing Helm
	// template data from the name is not necessary in other CLI commands).
	if anns := obj.GetAnnotations(); anns == nil {
		opsJSON = append(opsJSON, `{"op": "add", "path": "/metadata/annotations", "value": {}}`)
	}

	opsJSON = append(opsJSON, fmt.Sprintf(`{"op": "add", "path": "/metadata/annotations/helm-dump~1name", "value": "%s"}`, obj.GetName()))

	opsJSON = append(opsJSON, `{"op": "remove", "path": "/metadata/managedFields"}`)

	patchJSON := fmt.Sprintf("[%s]", strings.Join(opsJSON, ","))

	patch, err := jsonpatch.DecodePatch([]byte(patchJSON))
	if err != nil {
		return transform.PluginResponse{}, err
	}

	resp.Patches = patch

	return resp, nil
}
