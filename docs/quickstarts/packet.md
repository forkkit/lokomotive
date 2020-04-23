# Lokomotive Packet quickstart guide

## Contents

* [Introduction](#introduction)
* [Requirements](#requirements)
* [Step 1: Install lokoctl](#step-1-install-lokoctl)
* [Step 2: Set up a working directory](#step-2-set-up-a-working-directory)
* [Step 3: Set up credentials from environment variables](#step-3-set-up-credentials-from-environment-variables)
* [Step 4: Define cluster configuration](#step-4-define-cluster-configuration)
* [Step 5: Create Lokomotive cluster](#step-5-create-lokomotive-cluster)
* [Verification](#verification)
* [Cleanup](#cleanup)
* [Troubleshooting](#troubleshooting)
* [Conclusion](#conclusion)
* [Next steps](#next-steps)

## Introduction

This guide shows how to create a Lokomotive cluster on [Packet](https://www.packet.com/). By the
end of this guide, you'll have a basic Lokomotive cluster running on Packet.

Lokomotive runs on top of [Flatcar Container Linux](https://www.flatcar-linux.org/). This guide
uses the `stable` channel.

The guide uses [Amazon Route 53](https://aws.amazon.com/route53/) as a DNS provider.

## Requirements

* A Packet account with
  [local BGP](https://www.packet.com/developers/docs/network/advanced/local-and-global-bgp/)
  enabled.
* A Packet project ID.
* A Packet API key with access to the relevant project.
* An AWS account.
* An AWS
  [access key ID and secret](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html).
* An AWS Route 53 zone (can be a subdomain).
* An SSH key pair for accessing the cluster nodes.
* Terraform `v0.12.x` installed.
* [terraform-provider-ct](https://github.com/poseidon/terraform-provider-ct) `v0.5.0` installed.
* `kubectl` installed.

## Steps

### Step 1: Install lokoctl

`lokoctl` is the command-line interface for managing Lokomotive clusters.

Download the `lokoctl` binary for your platform:

```console
# Linux
wget https://github.com/kinvolk/lokomotive/releases/download/v0.1.0/lokoctl_0.1.0_linux_amd64.tar.gz

# macOS
wget https://github.com/kinvolk/lokomotive/releases/download/v0.1.0/lokoctl_0.1.0_darwin_amd64.tar.gz
```

Extract the binary and copy it to a place under your `$PATH`:

```console
tar zxvf lokoctl_0.1.0_linux_amd64.tar.gz
sudo cp lokoctl_0.1.0_linux_amd64/lokoctl /usr/local/bin
```

### Step 2: Create a cluster configuration

Create a directory for the cluster-related files and navigate to it:

```console
mkdir lokomotive-demo && cd lokomotive-demo
```

Create a file named `cluster.lokocfg` with the following contents:

```hcl
cluster "packet" {
  asset_dir        = "./assets"
  cluster_name     = "lokomotive-demo"

  dns {
    zone = "example.com"
    provider {
      route53 {
        zone_id = "X27BF86A3BNP80"
      }
    }
  }

  facility = "ams1"
  project_id = "89273817-4f44-4b41-9f0c-cb00bf538542"

  ssh_pubkeys       = ["ssh-rsa AAAA..."]
  management_cidrs  = ["0.0.0.0/0"]
  node_private_cidr = "10.0.0.0/8"

  controller_count = 1

  worker_pool "pool-1" {
    count     = 2
    node_type = "t1.small.x86"
  }
}
```

### Step 3: Deploy the cluster

Set up the Packet and AWS credentials in your shell:

```console
export PACKET_AUTH_TOKEN=k84jfL83kJF849B776Nle4L3980fake
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7FAKE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYFAKE
```

>NOTE: If you have the AWS CLI installed and configured for an AWS account, you can omit the
>`AWS_*` variables above. `lokoctl` follows the standard AWS authentication methods, which means it
>will use the `default` AWS CLI profile if no explicit credentials are specified. Similarly,
>environment variables such as `AWS_PROFILE` can be used to instruct `lokoctl` to use a specific
>AWS CLI profile for AWS authentication.

Add a private key corresponding to one of the public keys specified in `ssh_pubkeys` to your `ssh-agent`:

```bash
ssh-add ~/.ssh/id_rsa
ssh-add -L
```

Deploy the cluster:

```console
lokoctl cluster apply
```

The deployment process may take a couple of minutes. Upon successful completion, an output similar
to the following is shown:

```console
Your configurations are stored in ./assets

Now checking health and readiness of the cluster nodes ...

Node                             Ready    Reason          Message                            
                                                                                             
lokomotive-demo-controller-0       True     KubeletReady    kubelet is posting ready status    
lokomotive-demo-pool-1-worker-0    True     KubeletReady    kubelet is posting ready status    
lokomotive-demo-pool-1-worker-1    True     KubeletReady    kubelet is posting ready status    

Success - cluster is healthy and nodes are ready!
```

## Verification

Use the generated `kubeconfig` file to access the cluster:

```console
export KUBECONFIG=$(pwd)/assets/cluster-assets/auth/kubeconfig
kubectl get nodes
```

Sample output:

```console
NAME                            STATUS   ROLES    AGE   VERSION
lokomotive-demo-controller-0      Ready    <none>   33m   v1.17.4
lokomotive-demo-pool-1-worker-0   Ready    <none>   33m   v1.17.4
lokomotive-demo-pool-1-worker-1   Ready    <none>   33m   v1.17.4
```

## Using the cluster

At this point you should have access to a Lokomotive cluster and can use it to deploy applications.

If you don't have any Kubernetes experience, you can check out the [Kubernetes
Basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/deploy-app/deploy-intro/) tutorial.

>NOTE: Lokomotive uses a relatively restrictive Pod Security Policy by default. This policy
>disallows running containers as root. Refer to the
>[Pod Security Policy documentation](../concepts/securing-lokomotive-cluster.md#cluster-wide-pod-security-policy)
>for more details.

## Cleanup

To destroy the cluster, execute the following command:

```console
lokoctl cluster destroy --confirm
```

You can now safely delete the directory created for this guide if you no longer need it.

## Troubleshooting

In case the deployment fails, adding the `-v` flag to `lokoctl cluster apply` can provide useful
information.

### Stuck at "copy controller secrets"

```console
...
module.packet-lokomotive-demo.null_resource.copy-controller-secrets: Still creating... (8m30s elapsed)
module.packet-lokomotive-demo.null_resource.copy-controller-secrets: Still creating... (8m40s elapsed)
...
```

In case the deployment process seems to hang at the `copy-controller-secrets` phase for a long
time, check the following:

- Verify the correct private SSH key was added to `ssh-agent`.
- Verify that you can SSH into the created controller nodes from the machine running `lokoctl`.

### Packet provisioning failed

Sometimes the provisioning of servers on Packet may fail. In this case, retrying the deployment by
re-running `lokoctl cluster apply` may help.

### Insufficient capacity on Packet

Sometimes there may not be enough hardware available at a given Packet facility for a given machine
type. In this case, either select a different node type and/or Packet facility, or wait a while for
more capacity to become available. You can check the current capacity status on the Packet
[API](https://www.packet.com/developers/api/capacity/).

### Permission issues

If the deployment fails due to insufficient permissions on Packet, verify your Packet API key has
permissions to the right Packet project.

If the deployment fails due to insufficient permissions on AWS, ensure the IAM user associated with
the AWS API credentials has permissions to create records on Route 53.

## Next steps

**Lokomotive components** complement the "stock" Kubernetes functionality by adding features such
as load balancing, persistent storage and monitoring to a cluster. Refer to
[this](../concepts/components.md) document to learn about the available components as well as how
to deploy them.
