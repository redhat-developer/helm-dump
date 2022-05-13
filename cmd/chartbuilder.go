package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
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
	Logger        *logrus.Logger
	ProjectRoot   string
	OutputDir     string
	ResourceCache *ResourceCache
	Actions       []*Action
}

func NewChartBuilder(projectRoot string, outputDir string, logger *logrus.Logger) (*ChartBuilder, error) {
	return &ChartBuilder{
		Logger:      logger,
		ProjectRoot: projectRoot,
		OutputDir:   outputDir,
		ResourceCache: &ResourceCache{
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

type visitor struct {
	path        string
	beginOffset int
	endOffset   int
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	yamlPath := fmt.Sprintf("$%s", v.path)
	if node.GetPath() != yamlPath {
		return v
	}
	if n, ok := node.(*ast.MappingNode); !ok {
		return v
	} else {
		v.beginOffset = n.GetToken().Position.Offset
		v.endOffset = n.GetToken().Next.Position.Offset

		fmt.Printf("%s %s -> %d %d\n", n.GetPath(), n.GetToken().Type, v.beginOffset, v.endOffset)
	}
	return v
}

func updateTemplate(valuesKey string, path string, tmpl *chart.File) error {
	tmplAst, parseErr := parser.ParseBytes(tmpl.Data, 0)
	if parseErr != nil {
		return fmt.Errorf("error parsing template data: %w", parseErr)
	}

	v := &visitor{path: path}
	ast.Walk(v, tmplAst.Docs[0])

	fst := tmpl.Data[0 : v.beginOffset+1]
	snd := tmpl.Data[v.endOffset:len(tmpl.Data)]

	templateNewData := bytes.Join(
		[][]byte{
			fst,
			[]byte(fmt.Sprintf("{{ .Values.%s }}\n", valuesKey)),
			snd,
		},
		[]byte{})

	tmpl.Data = templateNewData

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

type ResourceCache struct {
	RootDir string
}

func (c *ResourceCache) GetResourcePath(key string) string {
	maybeResourcePath := filepath.Join(c.RootDir, replacer.Replace(key))
	return maybeResourcePath
}

func (c *ResourceCache) Exists(key string) (bool, error) {
	_, err := os.Stat(c.GetResourcePath(key))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *ResourceCache) Store(key string, data []byte) error {
	if _, err := os.Stat(c.RootDir); errors.Is(err, os.ErrNotExist) {
		mkdirErr := os.MkdirAll(c.RootDir, os.ModePerm)
		if mkdirErr != nil {
			return fmt.Errorf("error creating cache root dir: %w", mkdirErr)
		}
	}
	err := ioutil.WriteFile(c.GetResourcePath(key), data, 0644)
	if err != nil {
		return fmt.Errorf("error writing resource cache: %w", err)
	}
	return nil
}

func (c *ResourceCache) GetCachedResource(key string, data []byte) ([]byte, error) {
	exists, err := c.Exists(key)
	if err != nil {
		return nil, fmt.Errorf("error checking if cache key exists: %w", err)
	}
	if exists {
		cachedBytes, err := ioutil.ReadFile(c.GetResourcePath(key))
		if err != nil {
			return nil, fmt.Errorf("error reading cached resource: %w", err)
		}
		return cachedBytes, nil
	}

	var out map[string]interface{}
	unmarshalErr := yaml.Unmarshal(data, &out)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("error unmarshalling resource bytes: %w", unmarshalErr)
	}
	storeErr := c.Store(key, data)
	if storeErr != nil {
		return nil, storeErr
	}
	return data, nil
}

func (b *ChartBuilder) GetCachedResource(tmpl *chart.File) (*chart.File, error) {
	cachedBytes, err := b.ResourceCache.GetCachedResource(tmpl.Name, tmpl.Data)
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
