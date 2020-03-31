// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/backend/local"
	componentspkg "github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/config"
	"github.com/kinvolk/lokomotive/pkg/flatcar"
	"github.com/kinvolk/lokomotive/pkg/network"
	"github.com/kinvolk/lokomotive/pkg/platform"
)

// GetConfiguredCluster loads cluster from the given configuration file.
func GetConfiguredCluster(lokoConfig *config.Config) (Cluster, hcl.Diagnostics) {

	p, diags := getConfiguredPlatform(lokoConfig)
	if diags.HasErrors() {
		return nil, diags
	}

	if p == nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("no platform configured"),
		}
		return nil, hcl.Diagnostics{diag}
	}

	// Get the configured backend for the cluster. Backend types currently supported: local, s3.
	b, diags := getConfiguredBackend(lokoConfig)
	if diags.HasErrors() {
		return nil, diags
	}

	// Use a local backend if no backend is configured.
	if b == nil {
		b = local.NewLocalBackend()
	}

	// Get the configured flatcar for the cluster.
	fc, diags := getConfiguredFlatcar(lokoConfig)
	if diags.HasErrors() {
		return nil, diags
	}

	// Get the configured network for the cluster.
	n, diags := getConfiguredNetwork(lokoConfig)
	if diags.HasErrors() {
		return nil, diags
	}

	// Get the configured components for the cluster.
	c, diags := getConfiguredComponents(lokoConfig)
	if diags.HasErrors() {
		return nil, diags
	}

	clusterplatform, err := GetClusterPlatform(lokoConfig.ClusterConfig.Cluster.Name)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
		}
		return nil, hcl.Diagnostics{diag}
	}

	clusterplatform.SetFlatcar(fc)
	clusterplatform.SetPlatform(p)
	clusterplatform.SetBackend(b)
	clusterplatform.SetNetwork(n)
	clusterplatform.SetComponents(c)

	return clusterplatform, nil
}

// getConfiguredComponents loads components from the given configuration file.
func getConfiguredComponents(lokoConfig *config.Config) (map[string]componentspkg.Component, hcl.Diagnostics) {
	configuredComponents := map[string]componentspkg.Component{}
	for _, c := range lokoConfig.ClusterConfig.Components {
		componentConfigBody := lokoConfig.LoadComponentConfigBody(c.Name)
		component, err := componentspkg.Get(c.Name)
		if err != nil {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  err.Error(),
			}
			return map[string]componentspkg.Component{}, hcl.Diagnostics{diag}
		}
		diags := component.LoadConfig(componentConfigBody, lokoConfig.EvalContext)
		if diags.HasErrors() {
			return map[string]componentspkg.Component{}, diags
		}
		configuredComponents[c.Name] = component
	}

	return configuredComponents, hcl.Diagnostics{}
}

func ComponentsToApply(
	componentNames []string,
	configuredComponents map[string]componentspkg.Component,
) map[string]componentspkg.Component {
	componentsToApply := map[string]componentspkg.Component{}

	// if no component names are provided, then install all configured components
	if len(componentNames) == 0 {
		componentsToApply = configuredComponents
	}

	for _, name := range componentNames {
		c, ok := configuredComponents[name]
		if ok {
			componentsToApply[name] = c
		}
	}

	return componentsToApply
}

// getConfiguredFlatcar loads flatcar object from the given configuration file.
func getConfiguredFlatcar(lokoConfig *config.Config) (flatcar.Flatcar, hcl.Diagnostics) {
	if lokoConfig.ClusterConfig.Flatcar == nil {
		// No backend defined and no configuration error
		return nil, hcl.Diagnostics{}
	}

	fc, err := flatcar.GetFlatcar(lokoConfig.ClusterConfig.Cluster.Name)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
		}
		return nil, hcl.Diagnostics{diag}
	}

	return fc, fc.LoadConfig(&lokoConfig.ClusterConfig.Flatcar.Config, lokoConfig.EvalContext)
}

// getConfiguredNetwork loads network object from the given configuration file.
func getConfiguredNetwork(lokoConfig *config.Config) (network.Network, hcl.Diagnostics) {
	if lokoConfig.ClusterConfig.Network == nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Network not configured",
		}
		return nil, hcl.Diagnostics{diag}
	}

	n, err := network.GetNetwork(lokoConfig.ClusterConfig.Cluster.Name)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
		}
		return nil, hcl.Diagnostics{diag}
	}

	return n, n.LoadConfig(&lokoConfig.ClusterConfig.Network.Config, lokoConfig.EvalContext)
}

// getConfiguredBackend loads a backend from the given configuration file.
func getConfiguredBackend(lokoConfig *config.Config) (backend.Backend, hcl.Diagnostics) {
	if lokoConfig.ClusterConfig.Backend == nil {
		// No backend defined and no configuration error
		return nil, hcl.Diagnostics{}
	}

	backend, err := backend.GetBackend(lokoConfig.ClusterConfig.Backend.Name)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
		}
		return nil, hcl.Diagnostics{diag}
	}

	return backend, backend.LoadConfig(&lokoConfig.ClusterConfig.Backend.Config, lokoConfig.EvalContext)
}

// getConfiguredPlatform loads a platform from the given configuration file.
func getConfiguredPlatform(lokoConfig *config.Config) (platform.Platform, hcl.Diagnostics) {
	if lokoConfig.ClusterConfig.Cluster == nil {
		// No cluster defined and no configuration error
		return nil, hcl.Diagnostics{}
	}

	platform, err := platform.GetPlatform(lokoConfig.ClusterConfig.Cluster.Name)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
		}
		return nil, hcl.Diagnostics{diag}
	}

	return platform, platform.LoadConfig(&lokoConfig.ClusterConfig.Cluster.Config, lokoConfig.EvalContext)
}
