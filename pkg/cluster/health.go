package cluster

//import (
//	"fmt"
//	"os"
//	"text/tabwriter"
//
//	log "github.com/sirupsen/logrus"
//	"github.com/spf13/cobra"
//
//	"github.com/kinvolk/lokomotive/pkg/cluster"
//	"github.com/kinvolk/lokomotive/pkg/k8sutil"
//	"github.com/kinvolk/lokomotive/pkg/lokomotive"
//)
//
//func Health(cluster Cluster) error {
//
//	client, err := k8sutil.NewClientset(kubeconfig)
//	if err != nil {
//		contextLogger.Fatalf("Error in creating setting up Kubernetes client: %q", err)
//	}
//
//	lokomotivecluster, err := lokomotive.NewCluster(client, p.GetExpectedNodes())
//	if err != nil {
//		contextLogger.Fatalf("Error in creating new Lokomotive cluster: %q", err)
//	}
//
//	ns, err := lokomotivecluster.GetNodeStatus()
//	if err != nil {
//		contextLogger.Fatalf("Error getting node status: %q", err)
//	}
//
//	ns.PrettyPrint()
//
//	if !ns.Ready() {
//		contextLogger.Fatalf("The cluster is not completely ready.")
//	}
//
//	components, err := lokomotivecluster.Health()
//	if err != nil {
//		contextLogger.Fatalf("Error in getting Lokomotive cluster health: %q", err)
//	}
//
//	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
//
//	// Print the header.
//	fmt.Fprintln(w, "Name\tStatus\tMessage\tError\t")
//
//	// An empty line between header and the body.
//	fmt.Fprintln(w, "\t\t\t\t")
//
//	for _, component := range components {
//
//		// The client-go library defines only one `ComponenetConditionType` at the moment,
//		// which is `ComponentHealthy`. However, iterating over the list keeps this from
//		// breaking in case client-go adds another `ComponentConditionType`.
//		for _, condition := range component.Conditions {
//			line := fmt.Sprintf(
//				"%s\t%s\t%s\t%s\t",
//				component.Name, condition.Status, condition.Message, condition.Error,
//			)
//
//			fmt.Fprintln(w, line)
//		}
//
//		w.Flush()
//	}
//}
