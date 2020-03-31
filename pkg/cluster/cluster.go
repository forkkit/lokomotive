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

	//"github.com/hashicorp/hcl/v2"
	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/flatcar"
	"github.com/kinvolk/lokomotive/pkg/network"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

// Platform describes single environment, where cluster can be installed
type Cluster interface {
	Apply(*terraform.Executor) error
	ApplyComponents() error
	UpdateControlPlane(*terraform.Executor, bool) error
	Destroy(*terraform.Executor) error
	Initialize() error
	GetAssetDir() string
	Verify() error
	GetExpectedNodes() int
	SetPlatform(platform.Platform)
	SetFlatcar(flatcar.Flatcar)
	SetBackend(backend.Backend)
	SetNetwork(network.Network)
	SetComponents(map[string]components.Component)
	GetComponents() map[string]components.Component
}

// platforms is a collection where all platforms gets automatically registered
var clusterPlatforms map[string]Cluster

// initialize package's global variable when package is imported
func init() {
	clusterPlatforms = make(map[string]Cluster)
}

//Register adds platform into internal map
func Register(name string, c Cluster) {
	if _, exists := clusterPlatforms[name]; exists {
		panic(fmt.Sprintf("platform with name %q registered already", name))
	}
	clusterPlatforms[name] = c
}

// GetClusterPlatform returns platform based on the name
func GetClusterPlatform(name string) (Cluster, error) {
	clusterPlatform, exists := clusterPlatforms[name]
	if !exists {
		return nil, fmt.Errorf("no cluster platform with name %q found", name)
	}
	return clusterPlatform, nil
}
