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

package baremetal

var terraformConfigTmpl = `
module "bare-metal-{{.ClusterName}}" {
  source = "../lokomotive-kubernetes/bare-metal/flatcar-linux/kubernetes"

  providers = {
    local    = local.default
    null     = null.default
    template = template.default
    tls      = tls.default
  }

  # bare-metal
  cluster_name           = "{{.ClusterName}}"
  matchbox_http_endpoint = "{{.MatchboxHTTPEndpoint}}"
  os_channel             = "{{.OSChannel}}"
  os_version             = "{{.OSVersion}}"

  # configuration
  cached_install     = "{{.CachedInstall}}"
  k8s_domain_name    = "{{.K8sDomainName}}"
  ssh_keys           = {{.SSHPublicKeysRaw}}
  asset_dir          = "../cluster-assets"

  # machines
  controller_names   = {{.ControllerNamesRaw}}
  controller_macs    = {{.ControllerMacsRaw}}
  controller_domains = {{.ControllerDomainsRaw}}
  worker_names       = {{.WorkerNamesRaw}}
  worker_macs        = {{.WorkerMacsRaw}}
  worker_domains     = {{.WorkerDomainsRaw}}
}

provider "matchbox" {
  version     = "~> 0.3"
  endpoint    = "{{.MatchboxEndpoint}}"
  client_cert = file("{{.MatchboxClientCert}}")
  client_key  = file("{{.MatchboxClientKey}}")
  ca          = file("{{.MatchboxCA}}")
}

provider "ct" {
  version = "~> 0.3"
}

provider "local" {
  version = "1.4.0"
  alias   = "default"
}

provider "null" {
  version = "~> 2.1"
  alias   = "default"
}

provider "template" {
  version = "~> 2.1"
  alias   = "default"
}

provider "tls" {
  version = "~> 2.0"
  alias   = "default"
}

# Stub output, which indicates, that Terraform run at least once.
# Used when checking, if we should ask user for confirmation, when
# applying changes to the cluster.
output "initialized" {
  value = true
}

# values.yaml content for all deployed charts.
output "pod-checkpointer_values" {
  value = module.bare-metal-{{.ClusterName}}.pod-checkpointer_values
}

output "kube-apiserver_values" {
  value     = module.bare-metal-{{.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.bare-metal-{{.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.bare-metal-{{.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value = module.bare-metal-{{.ClusterName}}.calico_values
}
`
