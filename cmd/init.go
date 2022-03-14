package cmd

import (
	"context"
	"fmt"
	"helm.sh/helm/v3/pkg/chartutil"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/vmware-tanzu/velero/pkg/discovery"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/pager"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart"
)

func NewInitCmd() *cobra.Command {
	configFlags := genericclioptions.NewConfigFlags(true)

	var initCmd = &cobra.Command{
		Use:   "init chart-name output-dir",
		Short: "generates a Helm chart from existing resources",
		Long:  `Generates a Helm chart from existing resources`,
		Args:  cobra.ExactArgs(2),
		RunE:  buildInitCmd(configFlags),
	}

	configFlags.AddFlags(initCmd.Flags())

	return initCmd
}

func buildInitCmd(configFlags *genericclioptions.ConfigFlags) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		log := logrus.New()
		log.Level = logrus.DebugLevel

		discoveryClient, err := configFlags.ToDiscoveryClient()
		if err != nil {
			return err
		}

		discoveryClient.Invalidate()

		restConfig, err := configFlags.ToRESTConfig()
		if err != nil {
			return err
		}

		discoveryHelper, err := discovery.NewHelper(discoveryClient, log)
		if err != nil {
			return err
		}

		dynamicClient := dynamic.NewForConfigOrDie(restConfig)

		chartFiles := make([]*chart.File, 0)

		apiResourceLists := discoveryHelper.Resources()
		for _, resourceList := range apiResourceLists {
			gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
			if err != nil {
				continue
			}
			log.Debugf("Collecting definitions for %s", gv.String())
			for _, resource := range resourceList.APIResources {
				if !resource.Namespaced {
					continue
				}
				log.Debugf("\t%s.%s", gv.String(), resource.Kind)
				gvr := &schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: resource.Name,
				}
				resourceInterface := dynamicClient.Resource(*gvr)

				p := pager.New(func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
					return resourceInterface.Namespace(*configFlags.Namespace).List(ctx, opts)
				})

				list, _, err := p.List(cmd.Context(), metav1.ListOptions{})
				if err != nil {
					log.Errorf("%s", err)
					continue
				}

				err = meta.EachListItem(list, func(object runtime.Object) error {
					u, ok := object.(*unstructured.Unstructured)
					if !ok {
						return fmt.Errorf("expected *unstructured.Unstructured but got #{u}")
					}

					name := nameFromUnstructured(u)
					name = path.Join("templates", name)
					file, err := fileFromUnstructured(u, name)
					if err != nil {
						return err
					}

					chartFiles = append(chartFiles, file)

					return nil
				})
				if err != nil {
					log.Errorf("%s", err)
				}
			}
		}

		for _, chartFile := range chartFiles {
			log.Debugf("name: %s\ndata:\n%s", chartFile.Name, string(chartFile.Data))
		}

		name := args[0]
		chrt := &chart.Chart{
			Metadata: &chart.Metadata{
				APIVersion: "v2",
				Name:       name,
				Version:    "0.1.0",
			},
			Files: chartFiles,
		}

		outDir := args[1]
		save, err := chartutil.Save(chrt, outDir)
		if err != nil {
			return err
		}

		log.Debugf("chart stored in %s", save)

		return nil
	}
}

func fileFromUnstructured(obj *unstructured.Unstructured, name string) (*chart.File, error) {
	bytes, err := obj.MarshalJSON()
	if err != nil {
		return nil, err
	}

	bytes, err = yaml.JSONToYAML(bytes)
	if err != nil {
		return nil, err
	}

	file := &chart.File{
		Name: name,
		Data: bytes,
	}

	return file, nil
}

var replacer *strings.Replacer

func init() {
	replacer = strings.NewReplacer("/", "_", ".", "_")
}

func nameFromUnstructured(obj *unstructured.Unstructured) string {
	apiVersion := replacer.Replace(obj.GetAPIVersion())
	return fmt.Sprintf("%s_%s.yaml", obj.GetName(), apiVersion)
}

func init() {
	rootCmd.AddCommand(NewInitCmd())
}
