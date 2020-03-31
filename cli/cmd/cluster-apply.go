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

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokomotive/pkg/cluster"
	"github.com/kinvolk/lokomotive/pkg/install"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/lokomotive"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

var (
	verbose         bool
	skipComponents  bool
	upgradeKubelets bool
)

var clusterApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply configuration changes to a Lokomotive cluster with components",
	Run:   runClusterApply2,
}

func init() {
	clusterCmd.AddCommand(clusterApplyCmd)
	pf := clusterApplyCmd.PersistentFlags()
	pf.BoolVarP(&confirm, "confirm", "", false, "Upgrade cluster without asking for confirmation")
	pf.BoolVarP(&verbose, "verbose", "v", false, "Show output from Terraform")
	pf.BoolVarP(&skipComponents, "skip-components", "", false, "Skip applying component configuration")
	pf.BoolVarP(&upgradeKubelets, "upgrade-kubelets", "", false, "Experimentally upgrade self-hosted kubelets")
}

func runClusterApply2(cmd *cobra.Command, args []string) {

	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl cluster apply",
		"args":    args,
	})

	loko := initialize2(ctxLogger)

	// initialize platform
	loko.Initialize()
	ex, err := terraform.InitializeTerraform(loko.GetAssetDir(), verbose)
	if err != nil {
		ctxLogger.Fatalf("Failed to initialize terraform executor: %v", err)
	}

	exists, err := cluster.IsExists(ex)
	if err != nil {
		ctxLogger.Fatalf("Failed to check if the cluster exists: %v", err)
	}

	if exists && !confirm {
		// TODO: We could plan to a file and use it when installing.
		if err := ex.Plan(); err != nil {
			ctxLogger.Fatalf("Failed to reconcile cluster state: %v", err)
		}

		if !askForConfirmation("Do you want to proceed with cluster apply?") {
			ctxLogger.Println("Cluster apply cancelled")

			return
		}
	}

	if err = loko.Apply(ex); err != nil {
		ctxLogger.Fatalf("Failed to initialize cluster: %v", err)
	}

	if err = loko.Verify(); err != nil {
		ctxLogger.Fatalf("Unable to verify cluster: %v", err)
	}
	fmt.Printf("\nYour configurations are stored in %s\n", loko.GetAssetDir())

	// Do controlplane upgrades only if cluster already exists.
	if exists {
		loko.UpdateControlPlane(ex, upgradeKubelets)
		fmt.Printf("\nEnsuring that cluster controlplane is up to date.\n")
	}

	if skipComponents {
		return
	}

	//	componentsToApply := []string{}
	//	for _,name := loko.
	//	if err := loko.ApplyComponents(); err != nil {
	//		ctxLogger.Fatalf("Unable to install components: %v", err)
	//	}
}

//nolint:funlen
func runClusterApply(cmd *cobra.Command, args []string) {
	//
	//	componentsToApply := []string{}
	//	for _, component := range lokoConfig.ClusterConfig.Components {
	//		componentsToApply = append(componentsToApply, component.Name)
	//	}
	//
	//	ctxLogger.Println("Applying component configuration")
	//
	//	if len(componentsToApply) > 0 {
	//		if err := applyComponents(lokoConfig, kubeconfigPath, componentsToApply...); err != nil {
	//			ctxLogger.Fatalf("Applying component configuration failed: %v", err)
	//		}
	//	}
}

func verifyCluster(kubeconfigPath string, expectedNodes int) error {
	client, err := k8sutil.NewClientset(kubeconfigPath)
	if err != nil {
		return errors.Wrapf(err, "failed to set up clientset")
	}

	cluster, err := lokomotive.NewCluster(client, expectedNodes)
	if err != nil {
		return errors.Wrapf(err, "failed to set up cluster client")
	}

	return install.Verify(cluster)
}
