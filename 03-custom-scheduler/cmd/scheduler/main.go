package main

import (
	"github.com/fezho/k8s-examples/03-custom-scheduler/pkg/scheduler"
	"github.com/fezho/k8s-examples/03-custom-scheduler/pkg/signals"
	"k8s.io/api/core/v1"
	"k8s.io/klog"
)

func main() {
	klog.InitFlags(nil) // make the stderrThreshold value to info which is default

	klog.Info("starting custom scheduler")

	// TODO: use cache like kube-batch
	podQueue := make(chan *v1.Pod, 300)
	defer close(podQueue)

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	sched, err := scheduler.NewScheduler(podQueue)
	if err != nil {
		klog.Fatal(err)
	}
	sched.Run(stopCh)
}
