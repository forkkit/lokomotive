package flatcar

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
)

type Flatcar interface {

	// LoadConfig loads the flatcar config provided by the user.
	LoadConfig(*hcl.Body, *hcl.EvalContext) hcl.Diagnostics
	// Validate validates backend wnfiguration.
	Validate() error
}

// platforms is a collection where all platforms gets automatically registered
var platforms map[string]Flatcar

// initialize package's global variable when package is imported
func init() {
	platforms = make(map[string]Flatcar)
}

type Common struct {
	Channel string `hcl:"channel,optional"`
	Version string `hcl:"version,optional"`
}

// Register registers flatcar fc in the internal platforms map.
func Register(name string, fc Flatcar) {
	if _, exists := platforms[name]; exists {
		panic(fmt.Sprintf("platform with name %q registered already", name))
	}
	platforms[name] = fc
}

// GetFlatcar returns the Flatcar referred to by name.
func GetFlatcar(name string) (Flatcar, error) {
	flatcar, exists := platforms[name]
	if !exists {
		return nil, fmt.Errorf("no platform with name %q found", name)
	}
	return flatcar, nil
}
