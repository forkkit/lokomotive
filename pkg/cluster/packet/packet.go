package packet

import (
	"encoding/json"
	"fmt"
	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/cluster"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/dns"
	"github.com/kinvolk/lokomotive/pkg/flatcar"
	"github.com/kinvolk/lokomotive/pkg/install"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/lokomotive"
	"github.com/kinvolk/lokomotive/pkg/network"
	"github.com/kinvolk/lokomotive/pkg/platform"
	packetpkg "github.com/kinvolk/lokomotive/pkg/platform/packet"
	"github.com/kinvolk/lokomotive/pkg/platform/util"
	"github.com/kinvolk/lokomotive/pkg/terraform"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"sort"
	"text/template"
)

type packet struct {
	Platform   *packetpkg.Packet
	Backend    backend.Backend
	Flatcar    *PacketFlatcar
	Network    *PacketNetwork
	Components map[string]components.Component
}

// init registers packet as a platform
func init() {
	cluster.Register("packet", NewPacket())
}

func NewPacket() *packet {
	return &packet{}
}

func (p *packet) SetBackend(backend backend.Backend) {
	p.Backend = backend
}

func (p *packet) SetPlatform(packet platform.Platform) {
	p.Platform = packet.(*packetpkg.Packet)
}

func (p *packet) SetFlatcar(flatcar flatcar.Flatcar) {
	p.Flatcar = flatcar.(*PacketFlatcar)
}

func (p *packet) SetNetwork(network network.Network) {
	p.Network = network.(*PacketNetwork)
}

func (p *packet) SetComponents(components map[string]components.Component) {
	p.Components = components
}

func (p *packet) GetComponents() map[string]components.Component {
	return p.Components
}

func (p *packet) GetAssetDir() string {
	return p.Platform.Common.AssetDir
}

func (p *packet) ApplyComponents() error {

	return nil
}

func (p *packet) Initialize() error {
	if p.Platform.AuthToken == "" && os.Getenv("PACKET_AUTH_TOKEN") == "" {
		return fmt.Errorf("cannot find the Packet authentication token:\n" +
			"either specify AuthToken or use the PACKET_AUTH_TOKEN environment variable")
	}

	// TODO: Replace with p.Platform.GetAssetDir()
	// assetDir,err := p.Platform.GetAssetDir()
	assetDir, err := homedir.Expand(p.Platform.Common.AssetDir)
	if err != nil {
		return err
	}

	// Render backend configuration.
	renderedBackend, err := p.Backend.Render()
	if err != nil {
		return fmt.Errorf("Failed to render backend configuration file: %v", err)
	}

	// Configure Terraform directory, module and backend.
	if err := terraform.Configure(assetDir, renderedBackend); err != nil {
		return fmt.Errorf("Failed to configure Terraform : %v", err)
	}

	terraformRootDir := terraform.GetTerraformRootDir(assetDir)

	return createTerraformConfigFile(p, terraformRootDir)

}

func (p *packet) GetExpectedNodes() int {
	nodes := p.Platform.Common.ControllerCount
	for _, workerpool := range p.Platform.WorkerPools {
		nodes += workerpool.Count
	}

	return nodes
}

func (p *packet) Destroy(ex *terraform.Executor) error {

	fmt.Println("reached here to destroy ... ")
	if err := p.Initialize(); err != nil {
		return err
	}

	return ex.Destroy()
}

func (p *packet) Apply(ex *terraform.Executor) error {
	dnsProvider, err := dns.ParseDNS(&p.Platform.DNS)
	if err != nil {
		return errors.Wrap(err, "parsing DNS configuration failed")
	}

	if err := p.Initialize(); err != nil {
		return err
	}

	// Stop Execution for testing.
	return nil

	return p.terraformSmartApply(ex, dnsProvider)
}

func (p *packet) Verify() error {
	kubeconfig := cluster.GetKubeconfig(p.GetAssetDir())
	client, err := k8sutil.NewClientset(kubeconfig)
	if err != nil {
		return errors.Wrapf(err, "failed to set up clientset")
	}

	cluster, err := lokomotive.NewCluster(client, p.GetExpectedNodes())
	if err != nil {
		return errors.Wrapf(err, "failed to set up cluster client")
	}

	return install.Verify(cluster)
}

func createTerraformConfigFile(packet *packet, terraformPath string) error {
	tmplName := "cluster.tf"
	t := template.New(tmplName)
	t, err := t.Parse(terraformConfigTmpl)
	if err != nil {
		return errors.Wrap(err, "failed to parse template")
	}

	path := filepath.Join(terraformPath, tmplName)
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %q", path)
	}
	defer f.Close()

	keyListBytes, err := json.Marshal(packet.Platform.Common.SSHPubKeys)
	if err != nil {
		return errors.Wrap(err, "failed to marshal SSH public keys")
	}

	managementCIDRs, err := json.Marshal(packet.Network.ManagementCIDRs)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal management CIDRs")
	}

	// Packet does not accept tags as a key-value map but as an array of
	// strings.
	util.AppendTags(&packet.Platform.Common.Tags)
	tagsList := []string{}
	for k, v := range packet.Platform.Common.Tags {
		tagsList = append(tagsList, fmt.Sprintf("%s:%s", k, v))
	}
	sort.Strings(tagsList)
	tags, err := json.Marshal(tagsList)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal tags")
	}

	packet.Platform.Common.TagsRaw = string(tags)
	packet.Platform.Common.SSHPubKeysRaw = string(keyListBytes)
	packet.Network.ManagementCIDRsRaw = string(managementCIDRs)

	if err := t.Execute(f, packet); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}

	return nil
}

// terraformSmartApply applies cluster configuration.
func (p *packet) terraformSmartApply(ex *terraform.Executor, dnsProvider dns.DNSProvider) error {
	// If the provider isn't manual, apply everything in a single step.
	if dnsProvider != dns.DNSManual {
		return ex.Apply()
	}

	arguments := []string{"apply", "-auto-approve"}

	// Get DNS entries (it forces the creation of the controller nodes).
	arguments = append(arguments, fmt.Sprintf("-target=module.packet-%s.null_resource.dns_entries", p.Platform.Common.ClusterName))

	// Add worker nodes to speed things up.
	for _, w := range p.Platform.WorkerPools {
		arguments = append(arguments, fmt.Sprintf("-target=module.worker-%v.packet_device.nodes", w.Name))
	}

	// Create controller and workers nodes.
	if err := ex.Execute(arguments...); err != nil {
		return errors.Wrap(err, "failed executing Terraform")
	}

	if err := dns.AskToConfigure(ex, &p.Platform.DNS); err != nil {
		return errors.Wrap(err, "failed to configure DNS entries")
	}

	// Finish deployment.
	return ex.Apply()
}
