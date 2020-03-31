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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokomotive/pkg/cluster"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/config"
)

var componentApplyCmd = &cobra.Command{
	Use: "apply",
	Short: `Apply a component configuration. If not present it will install it.
If ran with no arguments it will apply all components mentioned in the
configuration.`,
	Run: runApply,
}

func init() {
	componentCmd.AddCommand(componentApplyCmd)
}

func runApply(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component apply",
		"args":    args,
	})

	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl cluster apply",
		"args":    args,
	})

	loko := initialize2(ctxLogger)

	var componentsToApply []string
	if len(args) == 0 {
		if err := loko.ApplyComponents(); err != nil {
			contextLogger.Fatalf("Unable to install configured components: %v", err)
		}
	} else {
		componentsToApply = append(componentsToApply, args...)
		components := cluster.ComponentsToApply(componentsToApply, loko.GetComponents())
		loko.SetComponents(components)
		if err := loko.ApplyComponents(); err != nil {
			contextLogger.Fatalf("Unable to install configured components: %v", err)
		}
	}
}

func applyComponents(lokoConfig *config.Config, kubeconfig string, componentNames ...string) error {
	for _, componentName := range componentNames {
		fmt.Printf("Applying component '%s'...\n", componentName)

		component, err := components.Get(componentName)
		if err != nil {
			return err
		}

		componentConfigBody := lokoConfig.LoadComponentConfigBody(componentName)

		if diags := component.LoadConfig(componentConfigBody, lokoConfig.EvalContext); len(diags) > 0 {
			fmt.Printf("%v\n", diags)
			return diags
		}

		if err := util.InstallComponent(componentName, component, kubeconfig); err != nil {
			return err
		}

		fmt.Printf("Successfully applied component '%s' configuration!\n", componentName)
	}
	return nil
}
