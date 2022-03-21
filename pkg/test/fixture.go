package test

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"testing"
)

func LoadYamlFixture(t *testing.T, path string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	serializer := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	bytes, err := ioutil.ReadFile(path)
	require.NoError(t, err)

	_, _, err = serializer.Decode(bytes, nil, obj)
	require.NoError(t, err)

	return obj
}
