package network

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
)

type Network interface {

	// LoadConfig loads the network config provided by the user.
	LoadConfig(*hcl.Body, *hcl.EvalContext) hcl.Diagnostics
	// Validate validates backend wnfiguration.
	Validate() error
}

// platforms is a collection where all platforms gets automatically registered
var platforms map[string]Network

// initialize package's global variable when package is imported
func init() {
	platforms = make(map[string]Network)
}

type Common struct {
	NetworkMTU          int    `hcl:"network_mtu,optional"`
	PodCIDR             string `hcl:"pod_cidr,optional"`
	ServiceCIDR         string `hcl:"service_cidr,optional"`
	ClusterDomainSuffix string `hcl:"cluster_domain_suffix,optional"`
	EnableReporting     bool   `hcl:"enable_reporting,optional"`
}

// Register registers network n in the internal platforms map.
func Register(name string, n Network) {
	if _, exists := platforms[name]; exists {
		panic(fmt.Sprintf("platform with name %q registered already", name))
	}
	platforms[name] = n
}

// GetNetwork returns the Network referred to by name.
func GetNetwork(name string) (Network, error) {
	network, exists := platforms[name]
	if !exists {
		return nil, fmt.Errorf("no platform with name %q found", name)
	}
	return network, nil
}
