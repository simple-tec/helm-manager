package util

import (
	"helm-manager/pkg/util/linux"
	"k8s.io/klog"
)

func GetHelm3CliPath() string {
	return "/bin/helm"
}

func AddHelmRepo(username string, passwd string, repoName string, url string) (bool, string) {
	// generate and run command
	helmPath := GetHelm3CliPath()
	str, err := linux.PostShCmd("chmod +x " + helmPath)
	if err != nil {
		klog.Errorf("AddHelmRepo error: %s\n", str)
	}

	command := GetHelm3CliPath() + " repo add --username " + username + " --password " + passwd + " " + repoName + " " + url + repoName
	klog.Info("command: ", command)
	str, err = linux.PostShCmd(command)
	if err != nil {
		klog.Errorf("AddHelmRepo error: %s\n", str)
		return false, ""
	} else {
		return true, repoName
	}
}

func InstallHelm(ns, repo, name string, chartName, versionName, kubeconfig, settings string) (string, error)  {
	var command string
	appinfo := repo + "/" + chartName + " --version " + versionName
	if settings == "" {
		command = GetHelm3CliPath() + " install " + name + " --namespace=" + ns + " " + appinfo + " --set --kubeconfig=" + kubeconfig
	} else {
		command = GetHelm3CliPath() + " install " + name + " --namespace=" + ns + " " + appinfo + " " + settings + " --set --kubeconfig=" + kubeconfig
	}

	klog.Info("command: ", command)
	str, err := linux.PostShCmd(command)
	return str, err
}

func RemoveHelm(ns, repo, name string, chartName, versionName, kubeconfig, settings string) (string, error)  {
	var command string
	//appinfo := repo + "/" + chartName + " --version " + versionName
	if settings == "" {
		command = GetHelm3CliPath() + " delete " + name + " --namespace=" + ns + " --kubeconfig=" + kubeconfig
	} else {
		command = GetHelm3CliPath() + " delete " + name + " --namespace=" + ns + " --kubeconfig=" + kubeconfig
	}

	klog.Info("command: ", command)
	str, err := linux.PostShCmd(command)
	return str, err
}
