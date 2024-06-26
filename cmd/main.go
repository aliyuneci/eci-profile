package main

import (
	"context"
	"flag"

	"eci.io/eci-profile/pkg/client/clientset/versioned"
	"eci.io/eci-profile/pkg/profile"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	var kubeConfig string
	var masterURL string
	var caCertPath string
	var caKeyPath string
	var qps float64
	var burst int
	flag.StringVar(&kubeConfig, "kubeconfig", "", "Path to a kubeConfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&caCertPath, "cacert", "", "Path to CA cert file in PEM format. Only for self-defined CA.")
	flag.StringVar(&caKeyPath, "cakey", "", "Path to CA key file in PEM format. Only for self-defined CA.")
	flag.Float64Var(&qps, "client-qps", 500, "k8s client maximum qps for throttle, default qps: 500")
	flag.IntVar(&burst, "client-burst", 1000, "k8s client maximum burst for throttle, default burst: 1000.")
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		klog.Fatalf("failed to build client config: %q", err)
	}
	cfg.QPS = float32(qps)
	cfg.Burst = burst
	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("failed to create client: %q", err)
	}
	profileClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("failed to create eci-profile client: %q", err)
	}

	profileConfig := &profile.Config{
		K8sClient:     k8sClient,
		ProfileClient: profileClient,
		CACertPath:    caKeyPath,
		CAKeyPath:     caKeyPath,
	}
	manager, err := profile.NewManager(profileConfig)
	if err != nil {
		klog.Fatalf("failed to create eci-profile manager: %q", err)
	}

	klog.Infof("ready to start eci-profile manager service")
	if err := manager.Run(context.TODO()); err != nil {
		klog.Fatalf("run profile service failed: %q", err)
	}
}
