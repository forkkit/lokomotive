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

package packet

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/kinvolk/lokomotive/pkg/common"
	"github.com/kinvolk/lokomotive/pkg/dns"
	"github.com/kinvolk/lokomotive/pkg/platform"
)

type workerPool struct {
	Name           string `hcl:"pool_name,label"`
	Count          int    `hcl:"count"`
	DisableBGP     bool   `hcl:"disable_bgp,optional"`
	IPXEScriptURL  string `hcl:"ipxe_script_url,optional"`
	OSArch         string `hcl:"os_arch,optional"`
	OSChannel      string `hcl:"os_channel,optional"`
	OSVersion      string `hcl:"os_version,optional"`
	NodeType       string `hcl:"node_type,optional"`
	Labels         string `hcl:"labels,optional"`
	Taints         string `hcl:"taints,optional"`
	SetupRaid      bool   `hcl:"setup_raid,optional"`
	SetupRaidHDD   bool   `hcl:"setup_raid_hdd,optional"`
	SetupRaidSSD   bool   `hcl:"setup_raid_ssd,optional"`
	SetupRaidSSDFS bool   `hcl:"setup_raid_ssd_fs,optional"`
}

type Packet struct {
	Common                common.CommonFields `hcl:",remain"`
	DNS                   dns.Config          `hcl:"dns,block"`
	Facility              string              `hcl:"facility"`
	ProjectID             string              `hcl:"project_id"`
	AuthToken             string              `hcl:"auth_token,optional"`
	ReservationIDs        map[string]string   `hcl:"reservation_ids,optional"`
	ReservationIDsDefault string              `hcl:"reservation_ids_default,optional"`
	WorkerPools           []workerPool        `hcl:"worker_pool,block"`
}

// init registers packet as a platform
func init() {
	platform.Register("packet", NewConfig())
}

func NewConfig() *Packet {
	return &Packet{
		Common: common.CommonFields{
			CertsValidityPeriodHours: 8760,
			ControllerCount:          1,
			ControllerType:           "baremetal_0",
			EnableAggregation:        true,
		},
	}
}

func (c *Packet) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	if diags := gohcl.DecodeBody(*configBody, evalContext, c); len(diags) != 0 {
		return diags
	}

	return c.checkValidConfig()
}

// GetAssetDir returns asset directory path
//func (c *Packet) GetAssetDir() string {
//	return c.Common.AssetDir
//}

// checkValidConfig validates cluster configuration.
func (c *Packet) checkValidConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	diagnostics = append(diagnostics, c.checkNotEmptyWorkers()...)
	diagnostics = append(diagnostics, c.checkWorkerPoolNamesUnique()...)

	return diagnostics
}

// checkNotEmptyWorkers checks if the cluster has at least 1 node pool defined.
func (c *Packet) checkNotEmptyWorkers() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if len(c.WorkerPools) == 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "At least one worker pool must be defined",
			Detail:   "Make sure to define at least one worker pool block in your cluster block",
		})
	}

	return diagnostics
}

// checkWorkerPoolNamesUnique verifies that all worker pool names are unique.
func (c *Packet) checkWorkerPoolNamesUnique() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	dup := make(map[string]bool)

	for _, w := range c.WorkerPools {
		if !dup[w.Name] {
			dup[w.Name] = true
			continue
		}

		// It is duplicated.
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Worker pools name should be unique",
			Detail:   fmt.Sprintf("Worker pool %v is duplicated", w.Name),
		})
	}

	return diagnostics
}
