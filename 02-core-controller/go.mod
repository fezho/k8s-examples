module github.com/fezho/k8s-examples/02-core-controller

go 1.12

require (
	github.com/fezho/k8s-examples/01-leader-election v0.0.0-20200203035835-886419de761c
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.3
	k8s.io/apimachinery v0.16.5-beta.1
	k8s.io/client-go v0.0.0-20191114101535-6c5935290e33
	k8s.io/klog v0.4.0
)
