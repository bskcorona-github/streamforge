package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bskcorona-github/streamforge/apps/operator/internal/controller"
	"github.com/bskcorona-github/streamforge/apps/operator/internal/config"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	metricsAddr          = flag.String("metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	probeAddr            = flag.String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	enableLeaderElection = flag.Bool("leader-elect", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	kubeconfig           = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		klog.Fatalf("Failed to load config: %v", err)
	}

	// Set up signals so we handle the first shutdown signal gracefully
	ctx := signals.SetupSignalHandler()

	// Get Kubernetes config
	k8sConfig, err := getKubernetesConfig()
	if err != nil {
		klog.Fatalf("Failed to get Kubernetes config: %v", err)
	}

	// Create manager
	mgr, err := manager.New(k8sConfig, manager.Options{
		Scheme:                 scheme.Scheme,
		MetricsBindAddress:     *metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: *probeAddr,
		LeaderElection:         *enableLeaderElection,
		LeaderElectionID:       "streamforge-operator",
	})
	if err != nil {
		klog.Fatalf("Failed to create manager: %v", err)
	}

	// Create controller
	if err := controller.AddToManager(mgr, cfg); err != nil {
		klog.Fatalf("Failed to add controller to manager: %v", err)
	}

	// Start manager
	klog.Info("Starting StreamForge Operator")
	if err := mgr.Start(ctx); err != nil {
		klog.Fatalf("Failed to start manager: %v", err)
	}
}

// getKubernetesConfig returns the Kubernetes configuration
func getKubernetesConfig() (*rest.Config, error) {
	if *kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}
	return rest.InClusterConfig()
} 