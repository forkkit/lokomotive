package packet

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/kinvolk/lokomotive/pkg/network"
	"net"
)

// init registers packet as a network for configuration.
func init() {
	network.Register("packet", NewPacketNetwork())
}

type PacketNetwork struct {
	network.Common     `hcl:",remain"`
	ManagementCIDRs    []string `hcl:"management_cidrs"`
	NodePrivateCIDR    string   `hcl:"node_private_cidr"`
	ManagementCIDRsRaw string
}

// LoadConfig loads the configuration for the s3 backend.
func (f *PacketNetwork) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, f)
}

func (p *PacketNetwork) Validate() error {
	if err := validCIDR(p.PodCIDR); err != nil {
		return fmt.Errorf("pod cidr '%s' not valid: %q", p.PodCIDR, err)
	}
	if err := validCIDR(p.ServiceCIDR); err != nil {
		return fmt.Errorf("service cidr '%s' not valid: %q", p.ServiceCIDR, err)
	}

	if p.NodePrivateCIDR != "" {
		if err := validCIDR(p.NodePrivateCIDR); err != nil {
			return fmt.Errorf("node private cidr '%s' not valid: %q", p.NodePrivateCIDR, err)
		}
	}

	for _, cidr := range p.ManagementCIDRs {
		if err := validCIDR(cidr); err != nil {
			return fmt.Errorf("management cidr '%s' not valid: %q", cidr, err)
		}
	}
	return nil
}

func validCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)

	return err
}

func NewPacketNetwork() *PacketNetwork {
	return &PacketNetwork{
		Common: network.Common{
			EnableReporting:     false,
			NetworkMTU:          1480,
			PodCIDR:             "10.2.0.0/16",
			ServiceCIDR:         "10.3.0.0/16",
			ClusterDomainSuffix: "cluster.local",
		},
		ManagementCIDRs:    []string{},
		NodePrivateCIDR:    "",
		ManagementCIDRsRaw: "",
	}
}
