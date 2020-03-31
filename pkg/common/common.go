package common

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
)

type Common interface {

	// LoadConfig loads the common config provided by the user.
	LoadConfig(*hcl.Body, *hcl.EvalContext) hcl.Diagnostics
	// Validate validates common configuration.
	Validate() error
}

// platforms is a collection where all platforms gets automatically registered
var platforms map[string]Common

// initialize package's global variable when package is imported
func init() {
	platforms = make(map[string]Common)
}

type CommonFields struct {
	AssetDir                 string            `hcl:"asset_dir"`
	ClusterName              string            `hcl:"cluster_name"`
	SSHPubKeys               []string          `hcl:"ssh_pubkeys"`
	Tags                     map[string]string `hcl:"tags,optional"`
	EnableAggregation        bool              `hcl:"enable_aggregation,optional"`
	CertsValidityPeriodHours int               `hcl:"certs_validity_period_hours,optional"`
	ControllerCount          int               `hcl:"controller_count"`
	ControllerType           string            `hcl:"controller_type,optional"`
	SSHPubKeysRaw            string
	TagsRaw                  string
}

// Register registers common c in the internal platforms map.
func Register(name string, c Common) {
	if _, exists := platforms[name]; exists {
		panic(fmt.Sprintf("platform with name %q registered already", name))
	}
	platforms[name] = c
}

// GetCommon returns the Common referred to by name.
func GetCommon(name string) (Common, error) {
	common, exists := platforms[name]
	if !exists {
		return nil, fmt.Errorf("no platform with name %q found", name)
	}
	return common, nil
}
