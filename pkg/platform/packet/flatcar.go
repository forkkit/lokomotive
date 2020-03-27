package packet

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/kinvolk/lokomotive/pkg/flatcar"
	"net/url"
)

// init registers packet as a flatcar for configuration.
func init() {
	flatcar.Register("packet", NewPacketFlatcar())
}

type PacketFlatcar struct {
	flatcar.Common `hcl:",remain"`
	Arch           string `hcl:"arch,optional"`
	//Channel       string `hcl:"channel,optional"`
	//Version       string `hcl:"version,optional"`
	IPXEScriptURL string `hcl:"ipxe_script_url,optional"`
}

// LoadConfig loads the configuration for the s3 backend.
func (f *PacketFlatcar) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, f)
}

func (f *PacketFlatcar) Validate() error {
	if f.Arch == "arm64" {
		if !isValidURL(f.IPXEScriptURL) {
			return fmt.Errorf("not a valid url: %s", f.IPXEScriptURL)
		}
	}

	return nil
}

func isValidURL(urlstring string) bool {
	_, err := url.ParseRequestURI(urlstring)
	if err != nil {
		return false
	}

	u, err := url.Parse(urlstring)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func NewPacketFlatcar() *PacketFlatcar {
	return &PacketFlatcar{
		Common: flatcar.Common{
			Version: "current",
			Channel: "stable",
		},
		Arch:          "amd64",
		IPXEScriptURL: "",
	}
}
