package cmd

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/redhat-developer/helm-dump/pkg/cache"
	"github.com/redhat-developer/helm-dump/pkg/visitor"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"

	_ "github.com/goccy/go-yaml"
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
	path       string
	template   string
}

type ChartBuilder struct {
	Logger      *logrus.Logger
	ProjectRoot string
	OutputDir   string
	Cache       *cache.Cache
	Actions     []*Action
}

func NewChartBuilder(projectRoot string, outputDir string, logger *logrus.Logger) (*ChartBuilder, error) {
	return &ChartBuilder{
		Logger:      logger,
		ProjectRoot: projectRoot,
		OutputDir:   outputDir,
		Cache: &cache.Cache{
			RootDir: filepath.Join(projectRoot, ".helm-dump"),
		},
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

func getJSONPathValue(obj map[string]interface{}, field string) (string, error) {
	j := jsonpath.New("")
	parseErr := j.Parse(fmt.Sprintf("{%s}", field))
	if parseErr != nil {
		return "", parseErr
	}
	execBuf := new(bytes.Buffer)
	execErr := j.Execute(execBuf, obj)
	if execErr != nil {
		return "", execErr
	}
	val := execBuf.String()
	return val, nil
}

func getResourceName(obj *unstructured.Unstructured) string {
	anns := obj.GetAnnotations()
	if anns == nil {
		return "unknown"
	}
	name, ok := anns["helm-dump/name"]
	if !ok {
		return "unknown"
	}
	return name
}

func renderActionTemplate(obj *unstructured.Unstructured, tmpl string) (string, error) {
	keyTmpl, err := template.New("").
		Funcs(map[string]interface{}{"resourceName": getResourceName}).
		Parse(tmpl)
	if err != nil {
		return "", err
	}
	keyBuf := new(bytes.Buffer)
	keyTmplErr := keyTmpl.Execute(keyBuf, obj)
	if keyTmplErr != nil {
		return "", keyTmplErr
	}
	key := keyBuf.String()
	return key, nil
}

func addToValuesYaml(
	obj *unstructured.Unstructured,
	valuesYaml map[string]interface{},
	path string,
	tmpl string,
) (string, error) {
	value, err := getJSONPathValue(obj.UnstructuredContent(), path)
	if err != nil {
		return "", err
	}

	key, renderErr := renderActionTemplate(obj, tmpl)
	if renderErr != nil {
		return "", err
	}

	setFieldErr := unstructured.SetNestedField(valuesYaml, value, strings.Split(key, ".")...)
	if setFieldErr != nil {
		return "", setFieldErr
	}

	return key, nil
}

func collectPatches(path string, node *ast.DocumentNode) []visitor.Patch {
	collector := visitor.NewCollector()
	v := visitor.NewMappingNodeVisitor(path, collector)
	ast.Walk(v, node)
	return collector.Patches
}

func updateTemplate(valuesKey string, path string, tmpl *chart.File) error {
	tmplAst, parseErr := parser.ParseBytes(tmpl.Data, 0)
	if parseErr != nil {
		return fmt.Errorf("error parsing template data: %w", parseErr)
	}

	patches := collectPatches(path, tmplAst.Docs[0])

	// Patches order must be descendent based on beginOffset.
	for _, p := range patches {
		tmpl.Data = p.Apply(valuesKey, tmpl.Data)
	}

	return nil
}

func appendValuesYaml(
	chrt *chart.Chart,
	valuesYaml map[string]interface{},
) error {
	// override chart values with the collected value
	valuesBytes, marshalErr := yaml.Marshal(valuesYaml)
	if marshalErr != nil {
		return fmt.Errorf("error marshalling values: %w", marshalErr)
	}

	// include values.yaml in the chart.
	chrt.Raw = append(
		chrt.Raw,
		&chart.File{
			Name: chartutil.ValuesfileName,
			Data: valuesBytes,
		},
	)

	return nil
}

func (b *ChartBuilder) GetCachedResource(tmpl *chart.File) (*chart.File, error) {
	cachedBytes, err := b.Cache.GetCachedResource(tmpl.Name, tmpl.Data)
	if err != nil {
		return nil, err
	}
	tmpl.Data = cachedBytes
	return tmpl, nil
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

	valuesYaml := make(map[string]interface{})

	// 2. process template resources that match apiVersion and kind.
TEMPLATE:
	for _, tmpl := range chrt.Templates {
		if isHiddenTemplate(tmpl) {
			continue TEMPLATE
		}

		if !isYAMLTemplate(tmpl) {
			continue TEMPLATE
		}

		tmpl, err := b.GetCachedResource(tmpl)
		if err != nil {
			b.Logger.WithError(err).Errorf("error obtaining cached resource")
			continue TEMPLATE
		}

		obj, gvk, decErr := decodeTemplate(tmpl)
		if decErr != nil {
			b.Logger.WithError(decErr).Errorf("error decoding template")
			continue TEMPLATE
		}

	ACTION:
		for _, action := range b.Actions {

			// only process apiVersion and kind specified in the command.
			if !actionMatchesGVK(action, gvk) {
				continue ACTION
			}

			valuesKey, err := addToValuesYaml(obj, valuesYaml, action.path, action.template)
			if err != nil {
				b.Logger.WithError(err).Errorf("error appending values.yaml")
				continue ACTION
			}

			// update the template object
			updateTemplateErr := updateTemplate(valuesKey, action.path, tmpl)
			if updateTemplateErr != nil {
				b.Logger.WithError(err).Errorf("error updating template resource")
				continue ACTION
			}
		}
	}

	appendValuesYamlErr := appendValuesYaml(chrt, valuesYaml)
	if appendValuesYamlErr != nil {
		return nil, appendValuesYamlErr
	}

	return chrt, nil
}
