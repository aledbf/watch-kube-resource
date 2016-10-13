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
	"log"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/watch"
)

// ConfigmapWatcher ...
type ConfigmapWatcher struct {
	kubeClient *unversioned.Client
	namespace  string
	name       string
	cmdRunner  CmdRunner
	done       chan bool
	w          watch.Interface
}

// NewConfigmapWatcher defines a watcher for Configmaps
func NewConfigmapWatcher(kubeClient *unversioned.Client, configmap string, cmdRunner CmdRunner) Watcher {
	namespace, name := parseNamespaceName(configmap)
	cmw := ConfigmapWatcher{
		name:       name,
		namespace:  namespace,
		kubeClient: kubeClient,
		cmdRunner:  cmdRunner,
		done:       make(chan bool, 1),
	}

	cmw.watch()
	return cmw
}

// OnEvent ...
func (cmw ConfigmapWatcher) OnEvent() {
	log.Printf("change in configmap %v/%v detected. Executing command...", cmw.namespace, cmw.name)
	stdout, stderr, err := cmw.cmdRunner.Run()
	if err != nil {
		log.Printf("error:\n%v\n", err)
	}
	if len(stdout) > 0 {
		log.Printf("command output:\n%v\n", stdout)
	}
	if len(stderr) > 0 {
		log.Printf("command error:\n%v\n", stderr)
	}
}

// Close ...
func (cmw ConfigmapWatcher) Close() error {
	cmw.done <- true
	cmw.w.Stop()
	return nil
}

// watch creates a fsnotify watcher for a file and create of write events
func (cmw *ConfigmapWatcher) watch() error {
	cm, err := cmw.kubeClient.ConfigMaps(cmw.namespace).Get(cmw.name)
	if err != nil {
		return err
	}

	sel := generic.ObjectMetaFieldsSet(cm.ObjectMeta, true)
	w, err := cmw.kubeClient.ConfigMaps(cmw.namespace).Watch(api.ListOptions{
		FieldSelector: sel.AsSelector(),
	})
	if err != nil {
		return err
	}
	cmw.w = w

	go func() {
		for {
			select {
			case event, ok := <-w.ResultChan():
				if !ok {
					return
				}
				if event.Type != watch.Added {
					cmw.OnEvent()
				}
			case <-cmw.done:
				log.Printf("configmap watcher done")
				return
			}
		}
	}()

	return nil
}
