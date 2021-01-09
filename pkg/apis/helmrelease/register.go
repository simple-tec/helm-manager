package helmrelease

import (
	extension "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crdClient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	GroupName = "simple.k8s.io"
	Version   = "v1"
	Kind      = "HelmRelease"
	Plural    = "helmreleases"
	ShortName = "helm"
)

func CreateCRD(client *crdClient.ApiextensionsV1beta1Client) {
	_, err := client.CustomResourceDefinitions().Get(Plural+"."+GroupName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Infof("%s CRD not existed,creating", Kind)
			helmreleaseCRD := extension.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: Plural + "." + GroupName,
				},
				Spec: extension.CustomResourceDefinitionSpec{
					Group:   GroupName,
					Version: Version,
					Names: extension.CustomResourceDefinitionNames{
						Kind:       Kind,
						Plural:     Plural,
						ShortNames: []string{ShortName},
					},
					Scope: extension.ResourceScope("Namespaced"),
					//Validation: validation,
					Subresources: &extension.CustomResourceSubresources{
						Status: &extension.CustomResourceSubresourceStatus{},
					},
				},
			}
			_, err = client.CustomResourceDefinitions().Create(&helmreleaseCRD)
			if err != nil {
				klog.Fatalf("Create crd %s error,errorMessage: %s", helmreleaseCRD.GetName(), err.Error())
			}
		} else {
			klog.Fatalf("Get crd %s error,errorMessage: %s", Plural+"."+GroupName, err.Error())
		}
	} else {
		klog.Infof("%s CRD existed,skip", Kind)
		return
	}
	klog.Infof("Create CRD %s done", Kind)
}
