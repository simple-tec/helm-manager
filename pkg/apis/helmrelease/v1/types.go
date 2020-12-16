package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status：

type HelmRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              HelmSpec   `json:"spec,omitempty"`
	Status            HelmStatus `json:"status,omitempty"`
}

type HelmSpec struct {
	RepoUrl string      `json:"repourl"`
	RepoName string      `json:"reponame"`
	User string  		`json:"user"`
	Password string 	`json:"password"`
	Namespace string  	`json:"namespace"`
	Name       string      `json:"name,omitempty"`
	Version        string      `json:"version,omitempty"`
	Settings      string      `json:"settings,omitempty"`
}

type HelmStatus struct {
	Status        string `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelmReleaseList is a list of HelmRelease resources
type HelmReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []HelmRelease `json:"items"`
}
