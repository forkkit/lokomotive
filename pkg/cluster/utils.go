package cluster

import (
	"github.com/kinvolk/lokomotive/pkg/terraform"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"path/filepath"
)

// IsExists determines if cluster has already been created by getting all
// outputs from the Terraform. If there is any output defined, it means 'terraform apply'
// run at least once.
func IsExists(ex *terraform.Executor) (bool, error) {
	o := map[string]interface{}{}

	if err := ex.Output("", &o); err != nil {
		return false, err
	}

	return (len(o) != 0), nil
}

// expandKubeconfigPath tries to expand ~ in the given kubeconfig path.
// However, if that fails, it just returns original path as the best effort.
func expandKubeconfigPath(path string) string {
	if expandedPath, err := homedir.Expand(path); err == nil {
		return expandedPath
	}

	// homedir.Expand is too restrictive for the ~ prefix,
	// i.e., it errors on "~somepath" which is a valid path,
	// so just return the original path.
	return path
}

// getKubeconfig finds the kubeconfig to be used. Precedence takes a specified
// flag or environment variable. Then the asset directory of the cluster is searched
// and finally the global default value is used. This cannot be done in Viper
// because we need the other values from Viper to find the asset directory.
func GetKubeconfig(assetDir string) string {
	kubeconfig := viper.GetString("kubeconfig")
	if kubeconfig != "" {
		return expandKubeconfigPath(kubeconfig)
	}
	if assetDir != "" {
		return expandKubeconfigPath(assetsKubeconfig(assetDir))
	}

	return expandKubeconfigPath("~/.kube/config")
}

func assetsKubeconfig(assetDir string) string {
	return filepath.Join(assetDir, "cluster-assets", "auth", "kubeconfig")
}
