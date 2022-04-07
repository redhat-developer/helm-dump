package main

import (
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/konveyor/crane-lib/transform"
	"github.com/redhat-developer/helm-dump/pkg/test"
	"github.com/stretchr/testify/require"
	"testing"
)

type OperationAsserter jsonpatch.Operation

func (m OperationAsserter) requirePath(t *testing.T, expected string) {
	p := jsonpatch.Operation(m)
	actual, err := p.Path()
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func (m OperationAsserter) requireValue(t *testing.T, expected interface{}) {
	p := jsonpatch.Operation(m)
	actual, err := p.ValueInterface()
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func (m OperationAsserter) requireKind(t *testing.T, expected string) {
	p := jsonpatch.Operation(m)
	require.Equal(t, expected, p.Kind())
}

func TestRun(t *testing.T) {

	t.Run("with-labels", func(t *testing.T) {
		// Arrange
		fixture := test.LoadYamlFixture(t, "test/test_run/nginx-deployment-with-labels.yaml")
		req := transform.PluginRequest{
			Unstructured: *fixture,
			Extras: map[string]string{
				"chart-name": "my-app",
			},
		}

		// Act
		resp, err := Run(req)
		require.NoError(t, err)

		// Assert
		require.Len(t, resp.Patches, 4)

		nameOp := OperationAsserter(resp.Patches[0])
		nameOp.requireKind(t, "add")
		nameOp.requirePath(t, "/metadata/name")
		nameOp.requireValue(t, "nginx-deployment-{{ .Release.Name }}")

		appNameOp := OperationAsserter(resp.Patches[1])
		appNameOp.requireKind(t, "add")
		appNameOp.requirePath(t, "/metadata/labels/app.kubernetes.io~1name")
		appNameOp.requireValue(t, `{{ template "my-app.fullname" $ }}`)

		appInstanceOp := OperationAsserter(resp.Patches[2])
		appInstanceOp.requireKind(t, "add")
		appInstanceOp.requirePath(t, "/metadata/labels/app.kubernetes.io~1instance")
		appInstanceOp.requireValue(t, "{{ $.Release.Name }}")
	})
	t.Run("without-labels", func(t *testing.T) {
		// Arrange
		fixture := test.LoadYamlFixture(t, "test/test_run/nginx-deployment-without-labels.yaml")
		req := transform.PluginRequest{
			Unstructured: *fixture,
			Extras: map[string]string{
				"chart-name": "my-app",
			},
		}

		// Act
		resp, err := Run(req)
		require.NoError(t, err)

		// Assert
		require.Len(t, resp.Patches, 5)

		nameOp := OperationAsserter(resp.Patches[0])
		nameOp.requireKind(t, "add")
		nameOp.requirePath(t, "/metadata/name")
		nameOp.requireValue(t, "nginx-deployment-{{ .Release.Name }}")

		labelsOp := OperationAsserter(resp.Patches[1])
		labelsOp.requireKind(t, "add")
		labelsOp.requirePath(t, "/metadata/labels")
		labelsOp.requireValue(t, map[string]interface{}{})

		appNameOp := OperationAsserter(resp.Patches[2])
		appNameOp.requireKind(t, "add")
		appNameOp.requirePath(t, "/metadata/labels/app.kubernetes.io~1name")
		appNameOp.requireValue(t, `{{ template "my-app.fullname" $ }}`)

		appInstanceOp := OperationAsserter(resp.Patches[3])
		appInstanceOp.requireKind(t, "add")
		appInstanceOp.requirePath(t, "/metadata/labels/app.kubernetes.io~1instance")
		appInstanceOp.requireValue(t, "{{ $.Release.Name }}")
	})

}
