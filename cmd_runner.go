/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	remotecommandserver "k8s.io/kubernetes/pkg/kubelet/server/remotecommand"
)

// CmdRunner ...
type CmdRunner interface {
	Run() (string, string, error)
}

// PodCmdRunner ...
type PodCmdRunner struct {
	kubeClient    *unversioned.Client
	clientConfig  *restclient.Config
	podName       string
	containerName string
	cmd           []string
}

// NewPodCmdRunner ...
func NewPodCmdRunner(kubeClient *unversioned.Client, clientConfig *restclient.Config,
	podName string, containerName string, cmd []string) CmdRunner {
	return PodCmdRunner{
		kubeClient:    kubeClient,
		clientConfig:  clientConfig,
		podName:       podName,
		containerName: containerName,
		cmd:           cmd,
	}
}

// Run ...
func (pcr PodCmdRunner) Run() (string, string, error) {
	var stdout, stderr bytes.Buffer
	name, namespace := parseNamespaceName(pcr.podName)
	req := pcr.kubeClient.RESTClient.Post().
		Resource("pods").
		Name(name).
		Namespace(namespace).
		SubResource("exec").
		Param("container", pcr.containerName)
	req.VersionedParams(&api.PodExecOptions{
		Container: pcr.containerName,
		Command:   pcr.cmd,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, api.ParameterCodec)

	exec, err := remotecommand.NewExecutor(pcr.clientConfig, "POST", req.URL())
	if err != nil {
		return "", "", err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		SupportedProtocols: remotecommandserver.SupportedStreamingProtocols,
		Stdout:             &stdout,
		Stderr:             &stderr,
		Tty:                false,
	})
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}
