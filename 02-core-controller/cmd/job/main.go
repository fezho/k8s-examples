package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/fezho/k8s-examples/01-leader-election/pkg/leaderelection"
	"github.com/fezho/k8s-examples/02-core-controller/pkg/controller"
	"github.com/fezho/k8s-examples/02-core-controller/pkg/signals"
	"github.com/spf13/pflag"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	kubemaster         = pflag.String("kubemaster", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	kubeconfig         = pflag.String("kubeconfig", "", "Absolute path to the kubeconfig file.")
	resyncPeriod       = pflag.Duration("resync-period", 60*time.Second, "resync period for job informer")
	retentionInSeconds = pflag.Int64("retention", 864000, "the retention period in seconds after job is completed, default value is 10 days")

	// for leader election
	enableLeaderElection = pflag.Bool("leader-elect", false, "whether to run the controller with leader election for high availability")
	podName              = pflag.String("holder-identity", os.Getenv("POD_NAME"), "the holder identity name")
	leaseLockName        = pflag.String("lease-lock-name", "", "the lease lock resource name")
	leaseLockNamespace   = pflag.String("lease-lock-namespace", os.Getenv("POD_NAMESPACE"), "the lease lock resource namespace")
)

func main() {
	klog.InitFlags(nil) // make the stderrThreshold value to info which is default

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	// context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// call cancel func either receiving shutdown signal or for the context itself to finish
		defer cancel()

		select {
		case <-stopCh:
		case <-ctx.Done():
		}
	}()

	// define the job controller running func
	run := func(ctx context.Context, kubeClient clientset.Interface) {
		sharedInformersFactory := informers.NewSharedInformerFactory(kubeClient, *resyncPeriod)
		jc := controller.NewJobController(kubeClient, sharedInformersFactory.Batch().V1().Jobs(), *retentionInSeconds)
		// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
		sharedInformersFactory.Start(ctx.Done())
		if err := jc.Run(1, ctx.Done()); err != nil {
			klog.Fatal("failed to run job controller, ", err)
		}
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags(*kubemaster, *kubeconfig)
	if err != nil {
		klog.Fatal("failed to build kubeconfig, ", err)
	}
	kubeClient := clientset.NewForConfigOrDie(kubeConfig)

	if *enableLeaderElection {
		klog.Info("run job controller with leader election")
		electionConfig := &leaderelection.Config{
			MemberID:           *podName,
			LeaseLockName:      *leaseLockName,
			LeaseLockNamespace: *leaseLockNamespace,
			LeaseDuration:      60 * time.Second, // the default value got error: failed to tryAcquireOrRenew context deadline exceeded
			RenewDeadline:      45 * time.Second,
			ComponentName:      "job-retention-controller",
			Callbacks: leaderelection.Callbacks{
				OnStartedLeading: func(ctx context.Context) {
					klog.Infof("%s: leading", *podName)
					run(ctx, kubeClient)
				},
				OnStoppedLeading: func() {
					//cancel()
					klog.Infof("%s: lost", *podName)
				},
			},
		}
		election, err := leaderelection.NewElectionWithKubeConfig(electionConfig, kubeConfig)
		if err != nil {
			klog.Fatalf("faled to init election, error: %v\n", err)
		}
		election.Run(ctx)
	} else {
		run(ctx, kubeClient)
	}

	klog.Info("job controller is stopped")
}
