package util

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string

	userId        int64 = 0
	userGroup     int64 = 0
	ensaasUserId  int64 = 1337
	ensaasGroupId int64 = 1337
	runAsNonRoot  bool  = false
	privileged    bool  = true
)


func GetK8SConfig() (*rest.Config, error) {
	mode := os.Getenv("mode")

	if mode == "production" {
		return rest.InClusterConfig()
	} else {
		flag.Parse()
		return clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	}
}

func Url2Resource(url string) (string, string, string, string, error) {
	kind := ""
	group := ""
	name := ""
	namespace := ""

	index := strings.Index(url, "?")
	if index > 0 {
		url = url[0:index]
	}

	if strings.HasPrefix(url, "/api/v1/") {
		urlArr := strings.Split(url, "/")
		switch len(urlArr) {
		case 4: // /api/v1/namespaces
			kind = urlArr[3]
		case 5: // /api/v1/namespaces/nsname
			kind = urlArr[3]
			name = urlArr[4]
		case 6: // /api/v1/namespaces/default/pods
			namespace = urlArr[4]
			kind = urlArr[5]
		case 7:
			// /api/v1/namespaces/default/pods/podname
			// /api/v1/namespaces/default/services/my-service
			namespace = urlArr[4]
			kind = urlArr[5]
			name = urlArr[6]
		case 8:
			namespace = urlArr[4]
			kind = urlArr[5]
			name = urlArr[6]
		default:
			return "", "", "", "", fmt.Errorf("Not found")
		}
	} else if strings.HasPrefix(url, "/apis/") {
		urlArr := strings.Split(url, "/")

		switch len(urlArr) {
		case 4:
			return "", "", "", "", nil
		case 5: // /apis/apps/v1/deployments    /apis/ensaas.k8s.io/v1/workspaces
			group = urlArr[2]
			kind = urlArr[4]
		case 6: // /apis/apps/v1/workspaces/wsname
			group = urlArr[2]
			kind = urlArr[4]
			name = urlArr[5]
		case 7: // /apis/apps/v1/namespaces/default/deployments
			group = urlArr[2]
			namespace = urlArr[5]
			kind = urlArr[6]
		case 8: //  /apis/apps/v1/namespaces/default/deployments/deployname
			group = urlArr[2]
			namespace = urlArr[5]
			kind = urlArr[6]
			name = urlArr[7]
		case 9: //  /apis/apps/v1/namespaces/default/deployments/deployname
			group = urlArr[2]
			namespace = urlArr[5]
			kind = urlArr[6]
			name = urlArr[7]
		default:
			return "", "", "", "", fmt.Errorf("Not found")
		}
	}

	return group, kind, namespace, name, nil
}

func Url2ResourceWithVersion(url string) (string, string, string, string, string, error) {
	kind := ""
	group := ""
	name := ""
	namespace := ""
	version := ""

	index := strings.Index(url, "?")
	if index > 0 {
		url = url[0:index]
	}

	if strings.HasPrefix(url, "/api/v1/") {
		urlArr := strings.Split(url, "/")
		switch len(urlArr) {
		case 4: // /api/v1/namespaces
			version = urlArr[2]
			kind = urlArr[3]
		case 5: // /api/v1/namespaces/nsname
			version = urlArr[2]
			kind = urlArr[3]
			name = urlArr[4]
		case 6: // /api/v1/namespaces/default/pods
			version = urlArr[2]
			namespace = urlArr[4]
			kind = urlArr[5]
		case 7: // /api/v1/namespaces/default/pods/podname
			version = urlArr[2]
			namespace = urlArr[4]
			kind = urlArr[5]
			name = urlArr[6]
		case 8:
			version = urlArr[2]
			namespace = urlArr[4]
			kind = urlArr[5]
			name = urlArr[6]
		default:
			return "", "", "", "", "", fmt.Errorf("Not found")
		}
	} else if strings.HasPrefix(url, "/apis/") {
		urlArr := strings.Split(url, "/")

		switch len(urlArr) {
		case 4:
			return "", "", "", "", "", nil
		case 5: // /apis/apps/v1/deployments    /apis/ensaas.k8s.io/v1/workspaces
			group = urlArr[2]
			version = urlArr[3]
			kind = urlArr[4]
		case 6: // /apis/apps/v1/workspaces/wsname
			group = urlArr[2]
			version = urlArr[3]
			kind = urlArr[4]
			name = urlArr[5]
		case 7: // /apis/apps/v1/namespaces/default/deployments
			group = urlArr[2]
			version = urlArr[3]
			namespace = urlArr[5]
			kind = urlArr[6]
		case 8: //  /apis/apps/v1/namespaces/default/deployments/deployname
			group = urlArr[2]
			version = urlArr[3]
			namespace = urlArr[5]
			kind = urlArr[6]
			name = urlArr[7]
		case 9: //  /apis/apps/v1/namespaces/default/deployments/deployname
			group = urlArr[2]
			version = urlArr[3]
			namespace = urlArr[5]
			kind = urlArr[6]
			name = urlArr[7]
		default:
			return "", "", "", "", "", fmt.Errorf("Not found")
		}
	}

	return group, kind, version, namespace, name, nil
}

func Config2String(cfg *rest.Config, userName, clusterName string) string {
	insecuryCertConfig :=
		`apiVersion: v1
kind: Config
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: %s
  name: %s
users:
- user:
    client-certificate-data: %s
    client-key-data: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s`

	certConfig :=
		`apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: %s
    server: %s
  name: %s
users:
- user:
    client-certificate-data: %s
    client-key-data: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s`

	insecuryTokenConfig :=
		`apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: %s
  name: %s
users:
- user:
    token: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s`

	tokenConfig :=
		`apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    certificate-authority-data: %s
    server: %s
  name: %s
users:
- user:
    token: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s`

	var config string
	if len(cfg.BearerToken) == 0 {
		if cfg.TLSClientConfig.CAData == nil {
			config = fmt.Sprintf(insecuryCertConfig, cfg.Host, clusterName, base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.CertData),
				base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.KeyData),
				userName, clusterName, userName, userName+"-"+clusterName, userName+"-"+clusterName)
		} else {
			config = fmt.Sprintf(certConfig,
				base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.CAData), cfg.Host,
				clusterName, base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.CertData),
				base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.KeyData),
				userName, clusterName, userName, userName+"-"+clusterName, userName+"-"+clusterName)
		}
	} else {
		if cfg.TLSClientConfig.CAData == nil {
			config = fmt.Sprintf(insecuryTokenConfig, cfg.Host, clusterName, cfg.BearerToken, userName, clusterName, userName,
				userName+"-"+clusterName, userName+"-"+clusterName)
		} else {
			config = fmt.Sprintf(tokenConfig, base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.CAData),
				cfg.Host, clusterName, cfg.BearerToken, userName, clusterName, userName,
				userName+"-"+clusterName, userName+"-"+clusterName)
		}
	}

	return config
}

func Config2StringWithNamespace(cfg *rest.Config, userName, clusterName, nsName string) string {
	insecuryCertConfig :=
		`apiVersion: v1
kind: Config
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: %s
  name: %s
users:
- user:
    client-certificate-data: %s
    client-key-data: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
    namespace: %s
  name: %s
current-context: %s`

	certConfig :=
		`apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: %s
    server: %s
  name: %s
users:
- user:
    client-certificate-data: %s
    client-key-data: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
    namespace: %s
  name: %s
current-context: %s`

	insecuryTokenConfig :=
		`apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: %s
  name: %s
users:
- user:
    token: %s
  name: %s
contexts:
- context:
    cluster: %s
	user: %s
    namespace: %s
  name: %s
current-context: %s`

	tokenConfig :=
		`apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    certificate-authority-data: %s
    server: %s
  name: %s
users:
- user:
    token: %s
  name: %s
contexts:
- context:
    cluster: %s
	user: %s
    namespace: %s
  name: %s
current-context: %s`

	var config string
	if len(cfg.BearerToken) == 0 {
		if cfg.TLSClientConfig.CAData == nil {
			config = fmt.Sprintf(insecuryCertConfig, cfg.Host, clusterName, base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.CertData),
				base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.KeyData),
				userName, clusterName, userName, nsName, userName+"-"+clusterName, userName+"-"+clusterName)
		} else {
			config = fmt.Sprintf(certConfig,
				base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.CAData), cfg.Host,
				clusterName, base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.CertData),
				base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.KeyData),
				userName, clusterName, userName, nsName, userName+"-"+clusterName, userName+"-"+clusterName)
		}
	} else {
		if cfg.TLSClientConfig.CAData == nil {
			config = fmt.Sprintf(insecuryTokenConfig, cfg.Host, clusterName, cfg.BearerToken, userName, clusterName, userName, nsName,
				userName+"-"+clusterName, userName+"-"+clusterName)
		} else {
			config = fmt.Sprintf(tokenConfig, base64.StdEncoding.EncodeToString(cfg.TLSClientConfig.CAData),
				cfg.Host, clusterName, cfg.BearerToken, userName, clusterName, userName, nsName,
				userName+"-"+clusterName, userName+"-"+clusterName)
		}
	}

	return config
}

func Token2Config(caCert, clusterHost, token, userName, clusterName string) string {
	insecuryTokenConfig :=
		`apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: %s
  name: %s
users:
- user:
    token: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s`

	tokenConfig :=
		`apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    certificate-authority-data: %s
    server: %s
  name: %s
users:
- user:
    token: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s`

	var config string

	if caCert == "" {
		config = fmt.Sprintf(insecuryTokenConfig, clusterHost, clusterName, token, userName, clusterName, userName,
			userName+"-"+clusterName, userName+"-"+clusterName)
	} else {
		config = fmt.Sprintf(tokenConfig, caCert,
			clusterHost, clusterName, token, userName, clusterName, userName,
			userName+"-"+clusterName, userName+"-"+clusterName)
	}

	return config
}

func Token2ConfigWithNamespace(caCert, clusterHost, token, userName, clusterName, nsName string) string {
	insecuryTokenConfig :=
		`apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: %s
  name: %s
users:
- user:
    token: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
    namespace: %s
  name: %s
current-context: %s`

	tokenConfig :=
		`apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    certificate-authority-data: %s
    server: %s
  name: %s
users:
- user:
    token: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
    namespace: %s
  name: %s
current-context: %s`

	var config string

	if caCert == "" {
		config = fmt.Sprintf(insecuryTokenConfig, clusterHost, clusterName, token, userName, clusterName, userName, nsName,
			userName+"-"+clusterName, userName+"-"+clusterName)
	} else {
		config = fmt.Sprintf(tokenConfig, caCert,
			clusterHost, clusterName, token, userName, clusterName, userName, nsName,
			userName+"-"+clusterName, userName+"-"+clusterName)
	}

	return config
}

func GetDomain() string {
	return "api-mp-ensaas." + os.Getenv("cluster") + "." + os.Getenv("domain")
}
func IsPodRunning(pod corev1.Pod) bool {
	result := true
	if pod.Status.ContainerStatuses == nil {
		if pod.Status.Phase != corev1.PodRunning {
			result = false
		}
	} else {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			result = result && containerStatus.Ready
		}
	}
	return result
}
