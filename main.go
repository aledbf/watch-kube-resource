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
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/pflag"

	"k8s.io/kubernetes/pkg/client/unversioned"
	kubectl_util "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func main() {
	flags := pflag.NewFlagSet("", pflag.ExitOnError)

	podName := flags.String("pod", "",
		`Name of the pod where the command will be executed. Takes the form namespace/name.
		No value means "this" pod (requires POD_NAME and POD_NAMESPACE using downward API) `)

	containerName := flags.String("container", "",
		`Name of hte container inside the pod. An empty string uses the first container in the pod`)

	file := flags.String("file", "", `Path to the file to watch`)

	configmap := flags.String("configmap", "", `Name of the configmap to watch. Takes the form namespace/name`)

	secret := flags.String("secret", "", `Name of the secret to watch. Takes the form namespace/name`)

	command := flags.String("command", "",
		`Path to the script or command to execute inside the pod`)

	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)
	clientConfig := kubectl_util.DefaultClientConfig(flags)

	if *command == "" {
		log.Fatalf("Please specify --command")
	}

	// if podname is empty check POD_NAME and POD_NAMESPACE env vars
	if *podName == "" {
		pod := os.Getenv("POD_NAME")
		if pod == "" {
			log.Fatal("You must specifiy the POD_NAME environment variable")
		}

		namespace := os.Getenv("POD_NAMESPACE")
		if namespace == "" {
			log.Fatal("You must specifiy the POD_NAMESPACE environment variable")
		}
		podns := fmt.Sprintf("%v/%v", pod, namespace)
		podName = &podns
	}

	activated := 0
	for _, aflag := range []string{*file, *configmap, *secret} {
		if aflag != "" {
			activated++
		}
	}
	if activated >= 2 {
		log.Fatalf("--file, --configmap and --secret options are mutually exclusive")
	}
	if activated == 0 {
		log.Fatalf("Please specify the flag --file, --configmap or --secret")
	}

	config, err := clientConfig.ClientConfig()
	if err != nil {
		log.Fatalf("error connecting to the client: %v", err)
	}
	kubeClient, err := unversioned.New(config)
	if err != nil {
		log.Fatalf("error connecting to the client: %v", err)
	}

	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	cmdRunner := NewPodCmdRunner(kubeClient, config, *podName, *containerName, strings.Split(*command, " "))
	// watch for the requested resource (file, secret, configmap, etc)
	var w Watcher
	if *file != "" {
		log.Printf("Watching file %v\n", *file)
		w, err = NewFileWatcher(*file, cmdRunner)
		if err != nil {
			log.Fatalf("error creating file watcher: %v", err)
		}
	} else if *configmap != "" {
		log.Printf("Watching configmap %v\n", *configmap)
		w = NewConfigmapWatcher(kubeClient, *configmap, cmdRunner)
	} else if *secret != "" {
		log.Printf("Watching secret %v\n", *secret)
		w = NewSecretWatcher(kubeClient, *secret, cmdRunner, done)
	}

	go func() {
		// wait for signal
		<-sigs
		// signal channel for end
		close(done)
		log.Println("Received SIGTERM, shutting down")
	}()
	// wait for end of execution
	<-done
	w.Close()
}

func parseNamespaceName(name string) (string, string) {
	parts := strings.Split(name, "/")
	if len(parts) == 1 {
		return "default", name
	}
	return parts[0], parts[1]
}
