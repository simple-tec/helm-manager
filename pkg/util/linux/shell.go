package linux

import (
	"bytes"
	"k8s.io/klog"
	"os/exec"
)

func PostBashCmd(command string) (string, error) {
	return exec_bash(command)
}

func exec_bash(s string) (string, error){
	cmd := exec.Command("/bin/bash", "-c", s)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}


func PostShCmd(command string) (string, error)  {
	return exec_sh(command)
}

func exec_sh(s string) (string, error){
	cmd := exec.Command("/bin/sh", "-c", s)
	klog.Infof("shell cmd: ", s)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	klog.Infof("shell stdout: ", out.String())
	klog.Infof("shell stderr: ", stderr.String())
	if err == nil {
		return out.String(), err
	} else {
		return stderr.String(), err
	}
}
