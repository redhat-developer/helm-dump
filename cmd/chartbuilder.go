package cmd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kyaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/util/jsonpath"
)

type Action struct {
	apiVersion string
	kind       string
	field      string
	template   string
}

type ChartBuilder struct {
	ProjectRoot string
	OutputDir   string
	Actions     []*Action
}

func NewChartBuilder(projectRoot string, outputDir string) (*ChartBuilder, error) {
	return &ChartBuilder{
		ProjectRoot: projectRoot,
		OutputDir:   outputDir,
	}, nil
}

func (b *ChartBuilder) AddAction(action *Action) {
	b.Actions = append(b.Actions, action)
}

func isHiddenTemplate(file *chart.File) bool {
	_, name := filepath.Split(file.Name)
	return strings.HasPrefix(name, "_")
}

func isYAMLTemplate(file *chart.File) bool {
	return strings.HasSuffix(file.Name, ".yaml") || strings.HasSuffix(file.Name, ".yml")
}

func decodeTemplate(file *chart.File) (*unstructured.Unstructured, *schema.GroupVersionKind, error) {
	dec := kyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	// decode the YAML template into unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, decErr := dec.Decode(file.Data, nil, obj)

	return obj, gvk, decErr
}

func actionMatchesGVK(action *Action, gvk *schema.GroupVersionKind) bool {
	actualApiVersion, actualKind := gvk.ToAPIVersionAndKind()
	return actualApiVersion == action.apiVersion && actualKind == action.kind
}

func (b *ChartBuilder) Build() (*chart.Chart, error) {
	chrt, loadErr := loader.LoadDir(b.ProjectRoot)
	if loadErr != nil {
		projectRoot, absErr := filepath.Abs(b.ProjectRoot)
		if absErr != nil {
			return nil, fmt.Errorf("error loading chart: %q might not be a directory; %w", b.ProjectRoot, absErr)
		}
		return nil, fmt.Errorf("error loading chart from %q: %w", projectRoot, loadErr)
	}

	values := make(map[string]interface{})

	// 2. process template resources that match apiVersion and kind.
TEMPLATE:
	for _, tmpl := range chrt.Templates {
		if isHiddenTemplate(tmpl) {
			continue TEMPLATE
		}

		if !isYAMLTemplate(tmpl) {
			continue TEMPLATE
		}

		obj, gvk, decErr := decodeTemplate(tmpl)
		if decErr != nil {
			continue TEMPLATE
		}

	ACTION:
		for _, action := range b.Actions {

			// only process apiVersion and kind specified in the command.
			if !actionMatchesGVK(action, gvk) {
				continue ACTION
			}

			// extract template value with jsonpath (TODO: refactor this into a function)
			j := jsonpath.New("")
			parseErr := j.Parse(fmt.Sprintf("{%s}", action.field))
			if parseErr != nil {
				continue ACTION
			}
			execBuf := new(bytes.Buffer)
			execErr := j.Execute(execBuf, obj.UnstructuredContent())
			if execErr != nil {
				continue ACTION
			}
			val := execBuf.String()

			// collect value to values.yaml
			keyTmpl, err := template.New("").Parse(action.template)
			if err != nil {
				continue ACTION
			}
			keyBuf := new(bytes.Buffer)
			keyTmplErr := keyTmpl.Execute(keyBuf, obj.UnstructuredContent())
			if keyTmplErr != nil {
				continue ACTION
			}
			key := keyBuf.String()

			// TODO: this is weak, should find a better way to collect the original resource name.
			key = strings.ReplaceAll(key, "-{{ .Release.Name }}", "")
			key = strings.ReplaceAll(key, "-", "_")
			unstructured.SetNestedField(values, val, strings.Split(key, ".")...)

			// update the template object
			newVal := fmt.Sprintf(`{{ .Values.%s }}`, key)
			fields := strings.Split(action.field, ".")
			if fields[0] == "" {
				fields = fields[1:]
			}
			setFieldErr := unstructured.SetNestedField(obj.Object, newVal, fields...)
			if setFieldErr != nil {
				continue ACTION
			}
			newData, marshalErr := yaml.Marshal(obj.Object)
			if marshalErr != nil {
				continue ACTION
			}
			tmpl.Data = newData
		}
	}

	// override chart values with the collected value
	valuesBytes, marshalErr := yaml.Marshal(values)
	if marshalErr != nil {
		return nil, fmt.Errorf("error marshalling values: %w", marshalErr)
	}

	// include values.yaml in the chart.
	chrt.Raw = append(
		chrt.Raw,
		&chart.File{
			Name: chartutil.ValuesfileName,
			Data: valuesBytes,
		},
	)

	return chrt, nil
}
