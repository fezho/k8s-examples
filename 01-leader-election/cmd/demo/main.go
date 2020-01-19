package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/fezho/k8s-examples/01-leader-election/pkg"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/leaderelection"
)

var (
	kubemaster         = pflag.String("kubemaster", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	kubeconfig         = pflag.String("kubeconfig", "", "Absolute path to the kubeconfig file.")
	podName            = pflag.String("holder-identity", os.Getenv("POD_NAME"), "the holder identity name")
	leaseLockName      = pflag.String("lease-lock-name", "", "the lease lock resource name")
	leaseLockNamespace = pflag.String("lease-lock-namespace", os.Getenv("POD_NAMESPACE"), "the lease lock resource namespace")
)

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	cfg := &pkg.Config{
		MemberID:           *podName,
		ComponentName:      "demo",
		LeaseLockName:      *leaseLockName,
		LeaseLockNamespace: *leaseLockNamespace,
		KubeMaster:         *kubemaster,
		KubeConfig:         *kubeconfig,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				log.Printf("[INFO] %s: started leading", *podName)
			},
			OnStoppedLeading: func() {
				log.Printf("[INFO] %s: stopped leading", *podName)
			},
			OnNewLeader: func(identity string) {
				log.Printf("[INFO] %s: new leader: %s", *podName, identity)
			},
		},
	}

	election, err := pkg.NewElection(cfg)
	if err != nil {
		log.Fatalf("faled to init election, error: %v\n", err)
	}

	election.Run(context.TODO())
	log.Print("lost lease")
}
