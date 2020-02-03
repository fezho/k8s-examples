package leaderelection

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	k8sleaderelection "k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

const (
	leaseDuration = 15 * time.Second
	renewDeadline = 10 * time.Second
	retryPeriod   = 2 * time.Second
)

type Election struct {
	config  *Config
	elector *k8sleaderelection.LeaderElector
}

type Config struct {
	// MemberID is the identifier of the member for leader election
	MemberID string
	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack.
	LeaseDuration time.Duration
	// RenewDeadline is the duration that the acting master will retry
	// refreshing leadership before giving up.
	RenewDeadline time.Duration
	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	RetryPeriod time.Duration
	// ComponentName is used as the group name for this leader election group. If you run
	// multiple leader elector instances within a service you probably want to differentiate
	// them by name.
	ComponentName string
	// LeaseLockName is name of the lease resource lock that is used for leader election.
	// You probably don't ever need to set this.
	LeaseLockName string
	// LeaseLockNamespace is the namespace of the lease resource lock
	LeaseLockNamespace string
	// The address of the Kubernetes API server (overrides any value in kubeconfig)
	KubeMaster string
	// Path to kubeconfig. If left unset the in cluster config is used by default.
	KubeConfig string
	// Callbacks are callbacks that are triggered during certain lifecycle
	// events of the LeaderElector
	Callbacks Callbacks
}

type Callbacks struct {
	// OnStartedLeading is called when a LeaderElector client starts leading
	OnStartedLeading func(context.Context)
	// OnStoppedLeading is called when a LeaderElector client stops leading. Actually it is called when every member dies.
	OnStoppedLeading func()
	// OnNewLeader is called when the client observes a leader that is
	// not the previously observed leader. This includes the first observed
	// leader when the client starts.
	OnNewLeader func(identity string)
}

func NewElection(config *Config) (*Election, error) {
	kubeConfig, err := buildConfig(config.KubeMaster, config.KubeConfig)
	if err != nil {
		return nil, err
	}

	return NewElectionWithKubeConfig(config, kubeConfig)
}

func NewElectionWithKubeConfig(config *Config, kubeConfig *rest.Config) (*Election, error) {
	err := checkAndSetConfig(config)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restclient.AddUserAgent(kubeConfig, "leader-election"))
	if err != nil {
		return nil, err
	}

	// Prepare event clients.
	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: clientset.CoreV1().Events(config.LeaseLockNamespace)})
	eventRecorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: config.ComponentName})

	// Prepare resource lock
	rl, err := resourcelock.New(
		resourcelock.LeasesResourceLock,
		config.LeaseLockNamespace,
		config.LeaseLockName,
		clientset.CoreV1(),
		clientset.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      config.MemberID,
			EventRecorder: eventRecorder,
		})
	if err != nil {
		return nil, fmt.Errorf("couldn't create resource lock: %v", err)
	}

	// Prepare elector
	le, err := k8sleaderelection.NewLeaderElector(k8sleaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: config.LeaseDuration,
		RenewDeadline: config.RenewDeadline,
		RetryPeriod:   config.RetryPeriod,
		Callbacks: k8sleaderelection.LeaderCallbacks{
			OnStartedLeading: config.Callbacks.OnStartedLeading,
			OnStoppedLeading: config.Callbacks.OnStoppedLeading,
			OnNewLeader:      config.Callbacks.OnNewLeader,
		},
		ReleaseOnCancel: true,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't create leader elector: %v", err)
	}

	return &Election{
		config:  config,
		elector: le,
	}, nil
}

func (e *Election) Run(ctx context.Context) {
	e.elector.Run(ctx)
}

func (e *Election) IsLeader() bool {
	return e.elector.IsLeader()
}

func buildConfig(master, kubeconfig string) (*rest.Config, error) {
	if master != "" || kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(master, kubeconfig)
	}
	return rest.InClusterConfig()
}

func checkAndSetConfig(cfg *Config) error {
	if cfg.LeaseLockNamespace == "" {
		return fmt.Errorf("lease lock namespace must not be nil for leader election")
	}

	if cfg.MemberID == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("unable to get hostname: %v", err)
		}
		// add a uniquifier so that two processes on the same host don't accidentally both become active
		cfg.MemberID = hostname + "_" + string(uuid.NewUUID())
	}
	if cfg.LeaseLockName == "" {
		cfg.LeaseLockName = "leader-election"
	}
	if cfg.ComponentName == "" {
		cfg.ComponentName = "leader-election"
	}
	if cfg.LeaseDuration == 0 {
		cfg.LeaseDuration = leaseDuration
	}
	if cfg.RenewDeadline == 0 {
		cfg.RenewDeadline = renewDeadline
	}
	if cfg.RetryPeriod == 0 {
		cfg.RetryPeriod = retryPeriod
	}

	return nil
}
