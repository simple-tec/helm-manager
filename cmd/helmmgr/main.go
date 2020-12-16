package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"helm-manager/pkg/apis/helmrelease"
	clientset "helm-manager/pkg/client/helmrelease/clientset/versioned"
	informers "helm-manager/pkg/client/helmrelease/informers/externalversions"
	"helm-manager/pkg/controller"
	"helm-manager/pkg/signals"
	crdClient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	flag.Parse()
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	crdClient, err := crdClient.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building crd clientset: %s", err.Error())
	}
	helmrelease.CreateCRD(crdClient)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	helmClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}

	helmInformerFactory := informers.NewSharedInformerFactory(helmClient, time.Second*30)

    //得到controller
	controller := controller.NewHelmreleaseController(kubeClient, helmClient,
		helmInformerFactory.Helmrelease().V1().HelmReleases())

    //启动informer
	go helmInformerFactory.Start(stopCh)

    //controller开始处理消息
	if err = controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
