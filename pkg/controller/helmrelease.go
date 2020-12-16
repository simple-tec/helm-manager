package controller

import (
	"helm-manager/pkg/util"
	"fmt"
	"os"
	"runtime/debug"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"

	helmreleasev1 "helm-manager/pkg/apis/helmrelease/v1"
	helmClientSet "helm-manager/pkg/client/helmrelease/clientset/versioned"
	helmScheme "helm-manager/pkg/client/helmrelease/clientset/versioned/scheme"
	helmInformer "helm-manager/pkg/client/helmrelease/informers/externalversions/helmrelease/v1"
	helmListers "helm-manager/pkg/client/helmrelease/listers/helmrelease/v1"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const helmreleaseControllerAgentName = "helm-controller"
const (
	SuccessSynced         = "Synced"
	MessageResourceSynced = "Helm synced successfully"
	KubeConfig            = "/tmp/config"
)

type HelmreleaseController struct {
	kubeClientSet kubernetes.Interface
	helmClient    helmClientSet.Interface
	helmSynced    cache.InformerSynced
	helmLister    helmListers.HelmReleaseLister
	workqueue     workqueue.RateLimitingInterface
	recorder      record.EventRecorder
}

func saveConfigfile(file string, content string) bool {
	if content == "" {
		klog.Error("getAndWriteKubeConfig fail")
		return false
	}
	f, err := os.Create(file)
	if err != nil {
		fmt.Println(err)
		return false
	}
	l, err := f.WriteString(content)
	if err != nil {
		klog.Error(err)
		f.Close()
		return false
	}
	if l != len(content) {
		klog.Error("wirted, write:", l, len(content))
		f.Close()
		return false
	}
	f.Close()
	return true
}

func installHelm(namespace string, repoName string, repoUrl string, user string, passwd string, name string, charName string, charVersion string, settings string) (bool, string) {
	start := time.Now()
	defer func() {
		cost := time.Since(start)
		fmt.Println("installHelm=", cost)
	}()

	ret, str := util.AddHelmRepo(user, passwd, repoName, repoUrl)
	if ret == false {
		klog.Errorf("AddHelmRepo failed, error: %s\n", str)
		return false, "AddHelmRepo failed"
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			klog.Errorf("Fail to get config file")
			return false, "Fail to get config file"
		}
		cluster := os.Getenv("cluster")
		configStr := util.Config2String(config, "admin", cluster)
		//klog.Infof("config:%s\n", configStr)
		ok := saveConfigfile(KubeConfig, configStr)
		if ok == false {
			klog.Info("Save config failed\n")
			return false, "Save config failed"
		}
		str, err := util.InstallHelm(namespace, repoName, name, charName, charVersion, KubeConfig, settings)
		if err == nil {
			klog.Info("Install helm ok\n")
			return true, str
		} else {
			klog.Info("Install helm failed\n")
			return false, str
		}
	}
}

func removeHelm(namespace string, repoName string, repoUrl string, user string, passwd string, name string, charName string, charVersion string, settings string) (bool, string) {
	start := time.Now()
	defer func() {
		cost := time.Since(start)
		fmt.Println("removeHelm=", cost)
	}()

	ret, str := util.AddHelmRepo(user, passwd, repoName, repoUrl)
	if ret == false {
		klog.Errorf("AddHelmRepo failed, error: %s\n", str)
		return false, "AddHelmRepo failed"
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			klog.Errorf("Fail to get config file")
			return false, "Fail to get config file"
		}
		cluster := os.Getenv("cluster")
		configStr := util.Config2String(config, "admin", cluster)
		//klog.Infof("config:%s\n", configStr)
		ok := saveConfigfile(KubeConfig, configStr)
		if ok == false {
			klog.Info("Save config failed\n")
			return false, "Save config failed"
		}
		str, err := util.RemoveHelm(namespace, repoName, name, charName, charVersion, KubeConfig, settings)
		if err == nil {
			klog.Info("Remove helm ok\n")
			return true, str
		} else {
			klog.Info("Remove helm failed\n")
			return false, str
		}
	}
}

func NewHelmreleaseController(
	kubeClient kubernetes.Interface,
	helmClient helmClientSet.Interface,
	helmInformer helmInformer.HelmReleaseInformer) *HelmreleaseController {

	klog.Info("Creating event broadcaster by helm")

	utilruntime.Must(helmScheme.AddToScheme(scheme.Scheme))
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: helmreleaseControllerAgentName})

	controller := &HelmreleaseController{
		kubeClientSet: kubeClient,
		helmClient:    helmClient,
		helmLister:    helmInformer.Lister(),
		helmSynced:    helmInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "HelmReleases"),
		recorder:      recorder,
	}

	klog.Info("Setting up event handlers for helm")

	helmInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			go func(obj interface{}) {
				helmObj := obj.(*helmreleasev1.HelmRelease)
				klog.Info("AddFunc, %#v ...", helmObj)
				if helmObj.Status.Status != "" {
					return
				}
				namespace := helmObj.Spec.Namespace
				repoUrl := helmObj.Spec.RepoUrl
				user := helmObj.Spec.User
				password := helmObj.Spec.Password
				charName := helmObj.Spec.Name
				charVersion := helmObj.Spec.Version
				settings := helmObj.Spec.Settings
				repoName := helmObj.Spec.RepoName
				name := helmObj.Name
				//helmObj.SetFinalizers([]string{"finalizer.helmrelease"})
				klog.Infof("namespace:%s, url:%s, user:%s, passwd:%s, charName:%s, version:%s, settings:%s\n", namespace, repoUrl, user, password, charName, charVersion, settings)
				ret, str := installHelm(namespace, repoName, repoUrl, user, password, name, charName, charVersion, settings)
				if ret == true {
					helmObj.Status.Status = "Success"
				} else {
					helmObj.Status.Status = str
				}
				controller.helmClient.HelmreleaseV1().HelmReleases(namespace).UpdateStatus(helmObj)

			}(obj)
			return
		},
		UpdateFunc: func(old, new interface{}) {
			helmNew := new.(*helmreleasev1.HelmRelease)
			helmOld := old.(*helmreleasev1.HelmRelease)
			//klog.Info("UpdateFunc, %#v ...", helmNew)
			if helmNew.ResourceVersion == helmOld.ResourceVersion {
				return
			}
			//controller.enqueue(new)
		},
		DeleteFunc: func(obj interface{}) {
			go func(obj interface{}) {
				helmObj := obj.(*helmreleasev1.HelmRelease)
				klog.Info("DeleteFunc, %#v ...", helmObj)
				namespace := helmObj.Spec.Namespace
				repoUrl := helmObj.Spec.RepoUrl
				user := helmObj.Spec.User
				password := helmObj.Spec.Password
				charName := helmObj.Spec.Name
				charVersion := helmObj.Spec.Version
				settings := helmObj.Spec.Settings
				repoName := helmObj.Spec.RepoName
				name := helmObj.Name
				klog.Infof("namespace:%s, url:%s, user:%s, passwd:%s, charName:%s, version:%s, settings:%s\n", namespace, repoUrl, user, password, charName, charVersion, settings)
				removeHelm(namespace, repoName, repoUrl, user, password, name, charName, charVersion, settings)

			}(obj)
			return
		},
	})

	return controller
}

func (c *HelmreleaseController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	if ok := cache.WaitForCacheSync(stopCh, c.helmSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Start helm controller")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	klog.Info("Helm worker is started")
	<-stopCh
	klog.Info("Helm worker is stopped")
	return nil
}

func (c *HelmreleaseController) runWorker() {
	defer func() {
		err := recover()
		if err != nil {
			debug.PrintStack()
			klog.Info("testmessage err=", err)
		}
	}()
	for c.processNextWorkItem() {
	}
}

func (c *HelmreleaseController) processNextWorkItem() bool {

	obj, shutdown := c.workqueue.Get()

	if shutdown {
		klog.Info("Shut down the workqueue for helm")
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}

		c.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *HelmreleaseController) syncHandler(key string) error {

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	//get helm from cache
	klog.Infof("helm, namespace: %s, name: %s", namespace, name)
	helm, err := c.helmLister.HelmReleases(namespace).Get(name)
	if err != nil {
		// if obj is deleted ,will go here
		if errors.IsNotFound(err) {
			klog.Infof("helmrelease is deleted,  %s/%s ...", namespace, name)

			return nil
		}
		runtime.HandleError(fmt.Errorf("failed to list helmrelease by: %s/%s", namespace, name))
		return err
	}

	klog.Infof("helmrelease's expection: %#v ...", helm)

	c.recorder.Event(helm, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil

}

func (c *HelmreleaseController) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *HelmreleaseController) enqueueHelmForDelete(obj interface{}) {
	var key string
	var err error
	// delete obj from cache
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	// tick the key to listting event broadcaste
	c.workqueue.AddRateLimited(key)
}
