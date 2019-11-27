package contour

import (
	"testing"

	"github.com/hashicorp/hcl2/hcl"

	"github.com/kinvolk/lokoctl/pkg/components/util"
)

func TestEmptyConfig(t *testing.T) {
	c := newComponent()
	emptyConfig := hcl.EmptyBody()
	evalContext := hcl.EvalContext{}
	diagnostics := c.LoadConfig(&emptyConfig, &evalContext)
	if !diagnostics.HasErrors() {
		t.Fatalf("Empty config should return errors")
	}
}

func testRenderManifest(t *testing.T, configHCL string) {
	body, diagnostics := util.GetComponentBody(configHCL, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}

	component := newComponent()
	diagnostics = component.LoadConfig(body, &hcl.EvalContext{})
	if diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}

	m, err := component.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests with valid config should succeed, got: %s", err)
	}
	if len(m) <= 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}

func TestRenderManifest_WithInstallModeDeployment(t *testing.T) {
	configHCL := `
component "contour" {
  install_mode = "deployment"
}
`
	testRenderManifest(t, configHCL)
}

func TestRenderManifest_WithInstallModeDaemonSet(t *testing.T) {
	configHCL := `
component "contour" {
  install_mode = "daemonset"
}
`
	testRenderManifest(t, configHCL)
}

func TestRenderManifestWithInstallModeDeploymentAndServiceMonitor(t *testing.T) {
	configHCL := `
component "contour" {
  install_mode = "deployment"
  service_monitor = true
}
`
	testRenderManifest(t, configHCL)
}

func TestRenderManifestWithInstallModeDaemonSetAndServiceMonitor(t *testing.T) {
	configHCL := `
component "contour" {
  install_mode = "daemonset"
  service_monitor = true
}
`
	testRenderManifest(t, configHCL)
}
