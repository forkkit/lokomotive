package cluster

import (
	"fmt"
	"path/filepath"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

type controlplaneUpdater struct {
	kubeconfigPath string
	assetDir       string
	ex             *terraform.Executor
}

func (c *controlplaneUpdater) getControlplaneChart(name string) (*chart.Chart, error) {
	helmChart, err := loader.Load(filepath.Join(c.assetDir, "/lokomotive-kubernetes/bootkube/resources/charts", name))
	if err != nil {
		return nil, fmt.Errorf("loading chart from assets failed: %w", err)
	}

	if err := helmChart.Validate(); err != nil {
		return nil, fmt.Errorf("chart is invalid: %w", err)
	}

	return helmChart, nil
}

func (c *controlplaneUpdater) getControlplaneValues(name string) (map[string]interface{}, error) {
	valuesRaw := ""
	if err := c.ex.Output(fmt.Sprintf("%s_values", name), &valuesRaw); err != nil {
		return nil, fmt.Errorf("failed to get controlplane component values.yaml from Terraform: %w", err)
	}

	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(valuesRaw), &values); err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml for controlplane component: %w", err)
	}

	return values, nil
}

func UpdateControlPlane(ex *terraform.Executor, assetDir string, upgradeKubelets bool) error {
	cu := &controlplaneUpdater{
		kubeconfigPath: GetKubeconfig(assetDir),
		assetDir:       assetDir,
		ex:             ex,
	}

	releases := []string{"pod-checkpointer", "kube-apiserver", "kubernetes", "calico"}

	if upgradeKubelets {
		releases = append(releases, "kubelet")
	}

	for _, release := range releases {
		if err := cu.upgradeComponent(release); err != nil {
			return err
		}
	}

	return nil
}

func (c *controlplaneUpdater) upgradeComponent(component string) error {
	actionConfig, err := util.HelmActionConfig("kube-system", c.kubeconfigPath)
	if err != nil {
		return fmt.Errorf("Error updating control plane: failed initializing helm: %v", err)
	}

	helmChart, err := c.getControlplaneChart(component)
	if err != nil {
		return fmt.Errorf("Error updating control plane: loading chart '%s' from assets failed: %v", helmChart.Name(), err)
	}

	values, err := c.getControlplaneValues(component)
	if err != nil {
		return fmt.Errorf("Error updating control plane: failed to get kubernetes values.yaml from Terraform: %v", err)
	}

	exists, err := util.ReleaseExists(*actionConfig, component)
	if err != nil {
		return fmt.Errorf("Error updating control plane: failed to check component '%s' is installed: %v", component, err)
	}

	if !exists {
		fmt.Printf("Controlplane component '%s' is missing, reinstalling...", component)

		install := action.NewInstall(actionConfig)
		install.ReleaseName = component
		install.Namespace = "kube-system"
		install.Atomic = true

		if _, err := install.Run(helmChart, map[string]interface{}{}); err != nil {
			fmt.Println("Failed!")

			return fmt.Errorf("Installing controlplane component failed: %v", err)
		}

		fmt.Println("Done.")
	}

	update := action.NewUpgrade(actionConfig)

	update.Atomic = true

	fmt.Printf("Ensuring controlplane component '%s' is up to date... ", component)

	if _, err := update.Run(component, helmChart, values); err != nil {
		fmt.Println("Failed!")

		return fmt.Errorf("Updating chart failed: %v", err)
	}

	fmt.Println("Done.")

	return nil
}
