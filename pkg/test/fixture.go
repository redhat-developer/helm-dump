package test

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"testing"
)

var serializer runtime.Serializer

func init() {
	serializer = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
}

func LoadYamlFixture(t *testing.T, path string) *unstructured.Unstructured {
	bytes, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	return LoadBytesFixture(t, bytes)
}

func LoadBytesFixture(t *testing.T, bytes []byte) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	_, _, err := serializer.Decode(bytes, nil, obj)
	require.NoError(t, err)
	return obj
}
