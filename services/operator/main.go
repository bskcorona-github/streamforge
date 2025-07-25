package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	streamforgev1 "github.com/streamforge/streamforge/api/v1"
	"github.com/streamforge/streamforge/controllers"
	"github.com/streamforge/streamforge/services/operator/config"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(streamforgev1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var configPath string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&configPath, "config", "/etc/streamforge/operator/config.yaml", "Path to operator configuration file")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	})))

	// 設定の読み込み
	cfg, err := config.Load(configPath)
	if err != nil {
		setupLog.Error(err, "unable to load configuration")
		os.Exit(1)
	}

	// Kubernetes設定の取得
	kubeConfig, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "unable to get kubernetes config")
		os.Exit(1)
	}

	// マネージャーの作成
	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "streamforge-operator.streamforge.io",
		Cache: cache.Options{
			Namespaces: cfg.WatchNamespaces,
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// StreamProcessorコントローラーの設定
	if err = (&controllers.StreamProcessorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: cfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "StreamProcessor")
		os.Exit(1)
	}

	// MLEngineコントローラーの設定
	if err = (&controllers.MLEngineReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: cfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MLEngine")
		os.Exit(1)
	}

	// Collectorコントローラーの設定
	if err = (&controllers.CollectorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: cfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Collector")
		os.Exit(1)
	}

	// Webhookの設定
	if err = (&streamforgev1.StreamProcessor{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "StreamProcessor")
		os.Exit(1)
	}

	if err = (&streamforgev1.MLEngine{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "MLEngine")
		os.Exit(1)
	}

	if err = (&streamforgev1.Collector{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Collector")
		os.Exit(1)
	}

	// ヘルスチェックの設定
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
} 