package cmd

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kdiscovery "k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	ktesting "k8s.io/client-go/testing"

	hdtesting "github.com/redhat-developer/helm-dump/pkg/test"
)

var (
	ContextLines = 4
)

func newFakeCachedDiscovery(dynamicClient *fakedynamic.FakeDynamicClient) *FakeCachedDiscovery {
	discoveryClient := &FakeCachedDiscovery{
		FakeDiscovery: &fakediscovery.FakeDiscovery{
			Fake: &dynamicClient.Fake,
		},
	}

	discoveryClient.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: appsv1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{
					Name:       "deployments",
					Namespaced: true,
					Kind:       "Deployment",
					Group:      "apps",
					Version:    "v1",
					Verbs: []string{
						"list",
						"create",
						"get",
						"delete",
					},
				},
			},
		},
	}

	return discoveryClient
}

func makeInitCommandWorld(
	configFlags *genericclioptions.ConfigFlags,
	logger *logrus.Logger,
	objects ...runtime.Object,
) (*InitCommand, *FakeCachedDiscovery, *fakedynamic.FakeDynamicClient) {
	scheme := runtime.NewScheme()
	dynamicClient := fakedynamic.NewSimpleDynamicClient(scheme, objects...)
	discoveryClient := newFakeCachedDiscovery(dynamicClient)
	cmd, _ := NewInitCmd(configFlags, logger)
	cmd.PluginDir = "../plugins/helm_dump_init/dist/"
	cmd.DiscoveryClient = discoveryClient
	cmd.DynamicClient = dynamicClient
	return cmd, discoveryClient, dynamicClient
}

func TestNewInitCmd(t *testing.T) {

	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	t.Run("no-arguments", func(t *testing.T) {
		// Arrange
		configFlags := genericclioptions.NewConfigFlags(true)
		cmd, _ := NewInitCmd(configFlags, logger)
		cmd.SetArgs([]string{})

		// Act & Assert
		require.Error(t, cmd.Execute(), "Cmd requires arguments")
	})

	t.Run("minimum-required-arguments", func(t *testing.T) {
		// Arrange
		chartName := "my-chart"
		chartVersion := "0.1.0"

		tempDir := hdtesting.TempDir(t)

		cmd, discoveryClient, _ := makeInitCommandWorld(
			genericclioptions.NewConfigFlags(true),
			logger,
			hdtesting.LoadYamlFixture(t, "init_test/minimum-required-arguments/nginx-deployment.yaml"))

		cmd.SetArgs([]string{
			"--namespace", "default",
			chartName, tempDir})

		// Act
		require.NoError(t, cmd.Execute(), "Cmd must not return an error")

		// Assert
		expectedActions := []ktesting.Action{
			ktesting.NewListAction(
				schema.GroupVersionResource{Resource: "deployments", Group: "apps", Version: "v1"},
				schema.GroupVersionKind{Kind: "Deployment", Group: "apps", Version: "v1"},
				"default",
				metav1.ListOptions{},
			),
		}

		actualActions := FilterActions(discoveryClient.Actions())
		CheckActions(t, expectedActions, actualActions)

		hdtesting.RequireChartFileExists(t, tempDir, chartName, chartVersion)
		chrt := hdtesting.RequireChart(t, tempDir, chartName, chartVersion)

		require.Len(t, chrt.Templates, 2)

		maybeDeployment := chrt.Templates[0]
		actual := hdtesting.LoadBytesFixture(t, maybeDeployment.Data)
		require.Equal(t,
			"nginx-deployment-{{ .Release.Name }}",
			actual.GetName(),
			"name should match; did you build helm_dump_init crane plugin?")

		maybeHelpers := chrt.Templates[1]
		require.Equal(t, chartutil.HelpersName, maybeHelpers.Name)
	})

	t.Run("using-selector", func(t *testing.T) {
		// Arrange
		chartName := "my-chart"
		chartVersion := "0.1.0"
		labelSelector := "helm-dump=please"

		tempDir := hdtesting.TempDir(t)

		cmd, discoveryClient, _ := makeInitCommandWorld(
			genericclioptions.NewConfigFlags(true),
			logger,
			hdtesting.LoadYamlFixture(t, "init_test/using-selector/nginx-deployment1.yaml"),
			hdtesting.LoadYamlFixture(t, "init_test/using-selector/nginx-deployment2.yaml"),
		)

		cmd.SetArgs([]string{
			"--namespace", "default",
			"-l", labelSelector,
			chartName, tempDir})

		// Act
		require.NoError(t, cmd.Execute(), "Cmd must not return an error")

		// Assert
		expectedActions := []ktesting.Action{
			ktesting.NewListAction(
				schema.GroupVersionResource{Resource: "deployments", Group: "apps", Version: "v1"},
				schema.GroupVersionKind{Kind: "Deployment", Group: "apps", Version: "v1"},
				"default",
				metav1.ListOptions{LabelSelector: labelSelector},
			),
		}

		actualActions := FilterActions(discoveryClient.Actions())
		CheckActions(t, expectedActions, actualActions)

		expectedChartPath := path.Join(tempDir, fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))
		_, err := os.Stat(expectedChartPath)
		require.NoError(t, err, "%q should exist", expectedChartPath)

		chrt, err := loader.LoadFile(expectedChartPath)
		require.NoError(t, err, "%q should be a chart", expectedChartPath)

		require.Len(t, chrt.Templates, 2, "chart should contain %d templates, found %d", 2, len(chrt.Templates))

		maybeDeployment := chrt.Templates[0]
		actual := hdtesting.LoadBytesFixture(t, maybeDeployment.Data)
		require.Equal(t, "nginx-deployment1-{{ .Release.Name }}", actual.GetName(), "name should match; did you build helm_dump_init crane plugin?")

		maybeHelpers := chrt.Templates[1]
		require.Equal(t, chartutil.HelpersName, maybeHelpers.Name)
	})
}

type FakeCachedDiscovery struct {
	*fakediscovery.FakeDiscovery
}

func (f *FakeCachedDiscovery) Fresh() bool {
	return true
}

func (f *FakeCachedDiscovery) Invalidate() {
	// no-op
}

func CheckActions(t *testing.T, expected, actual []ktesting.Action) {
	for i, action := range actual {
		CheckAction(t, expected[i], action)
	}
}

func CheckAction(t *testing.T, expected, actual ktesting.Action) {
	if !equality.Semantic.DeepEqual(expected, actual) {
		diff, err := YamlDiff(expected, actual)
		if err != nil {
			panic(fmt.Sprintf("couldn't generate yaml diff: %s", err))
		}
		t.Errorf("expected action is different from actual:\n%s", diff)
	}
}

func FilterActions(actions []ktesting.Action) []ktesting.Action {
	filtered := make([]ktesting.Action, 0)
	for _, v := range actions {
		if ShouldSkip(v) {
			continue
		}
		filtered = append(filtered, v)
	}
	return filtered
}

func ShouldSkip(action ktesting.Action) bool {
	if action.GetResource().Resource == "group" ||
		action.GetResource().Resource == "resource" ||
		action.GetResource().Resource == "version" {
		return true
	}
	return false
}

func YamlDiff(expected interface{}, actual interface{}) (string, error) {
	yamlActual, err := yaml.Marshal(actual)
	if err != nil {
		return "", err
	}

	yamlExpected, err := yaml.Marshal(expected)
	if err != nil {
		return "", err
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(yamlExpected)),
		B:        difflib.SplitLines(string(yamlActual)),
		FromFile: "Expected",
		ToFile:   "Actual",
		Context:  ContextLines,
	}

	return difflib.GetUnifiedDiffString(diff)
}

// ChartFileExists ...
func ChartFileExists(
	t *testing.T,
	baseDir, chartName, chartVersion string,
) {
	expectedChartPath := path.Join(baseDir, fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))
	_, err := os.Stat(expectedChartPath)
	require.NoError(t, err, "%q should exist", expectedChartPath)
}

var _ kdiscovery.DiscoveryInterface = &FakeCachedDiscovery{}
var _ kdiscovery.CachedDiscoveryInterface = &FakeCachedDiscovery{}
