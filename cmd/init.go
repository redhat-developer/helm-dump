package cmd

import (
	"context"
	"fmt"
	"github.com/vmware-tanzu/velero/pkg/discovery"
	"helm.sh/helm/v3/pkg/chartutil"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kdiscovery "k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/pager"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	velerov1api "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/features"
	"helm.sh/helm/v3/pkg/chart"
)

type InitCommand struct {
	*cobra.Command
	Logger          *logrus.Logger
	DynamicClient   dynamic.Interface
	ConfigFlags     *genericclioptions.ConfigFlags
	DiscoveryClient kdiscovery.CachedDiscoveryInterface
	DiscoveryHelper discovery.Helper
}

func NewInitCmd(
	configFlags *genericclioptions.ConfigFlags,
	logger *logrus.Logger,
) (*InitCommand, error) {
	initCmd := &InitCommand{
		Command: &cobra.Command{
			Use:   "init chart-name output-dir",
			Short: "generates a Helm chart from existing resources",
			Long:  `Generates a Helm chart from existing resources`,
			Args:  cobra.ExactArgs(2),
		},
		Logger:      logger,
		ConfigFlags: configFlags,
	}
	initCmd.Command.RunE = initCmd.runE
	initCmd.Command.PreRunE = initCmd.preRunE

	configFlags.AddFlags(initCmd.Flags())

	return initCmd, nil
}

func (c *InitCommand) GetDiscoveryHelper() (discovery.Helper, error) {
	return discovery.NewHelper(c.DiscoveryClient, c.Logger)
}

func (c *InitCommand) preRunE(_ *cobra.Command, _ []string) error {
	if c.DynamicClient == nil {
		restConfig, err := c.ConfigFlags.ToRESTConfig()
		if err != nil {
			return err
		}
		c.DynamicClient = dynamic.NewForConfigOrDie(restConfig)
	}

	if c.DiscoveryClient == nil {
		var err error
		c.DiscoveryClient, err = c.ConfigFlags.ToDiscoveryClient()
		if err != nil {
			return err
		}
	}

	if c.DiscoveryHelper == nil {
		var err error
		c.DiscoveryHelper, err = discovery.NewHelper(c.DiscoveryClient, c.Logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *InitCommand) runE(cmd *cobra.Command, args []string) error {
	chartFiles := make([]*chart.File, 0)

	apiResourceLists := c.DiscoveryHelper.Resources()
	for _, resourceList := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			continue
		}
		c.Logger.Debugf("Collecting definitions for %s", gv.String())
		for _, resource := range resourceList.APIResources {
			if !resource.Namespaced {
				continue
			}
			c.Logger.Debugf("\t%s.%s", gv.String(), resource.Kind)
			gvr := &schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: resource.Name,
			}
			resourceInterface := c.DynamicClient.Resource(*gvr)

			p := pager.New(func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
				return resourceInterface.Namespace(*c.ConfigFlags.Namespace).List(ctx, opts)
			})

			list, _, err := p.List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				c.Logger.Errorf("%s", err)
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
				c.Logger.Errorf("%s", err)
			}
		}
	}

	for _, chartFile := range chartFiles {
		c.Logger.Debugf("name: %s\ndata:\n%s", chartFile.Name, string(chartFile.Data))
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

	c.Logger.Debugf("chart stored in %s", save)

	return nil

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

var replacer = strings.NewReplacer("/", "_", ".", "_")

func nameFromUnstructured(obj *unstructured.Unstructured) string {
	apiVersion := replacer.Replace(obj.GetAPIVersion())
	return fmt.Sprintf("%s_%s.yaml", obj.GetName(), apiVersion)
}

func init() {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	// required for the discovery.Helper to pull all api groups from the api server
	features.Enable(velerov1api.APIGroupVersionsFeatureFlag)

	configFlags := genericclioptions.NewConfigFlags(true)
	initCmd, err := NewInitCmd(configFlags, logger)
	if err != nil {
		panic(err)
	}

	rootCmd.AddCommand(initCmd.Command)
}