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
	"path"

	"gopkg.in/fsnotify.v1"
)

// Watcher ...
type Watcher interface {
	// OnEvent is the callback to be executed when the resource being watched changes
	OnEvent()
	// Close closes the watcher
	Close() error
}

// FileWatcher ...
type FileWatcher struct {
	file      string
	watcher   *fsnotify.Watcher
	cmdRunner CmdRunner
	done      chan bool
}

// NewFileWatcher defines a watcher for files
func NewFileWatcher(file string, cmdRunner CmdRunner) (Watcher, error) {
	fw := FileWatcher{
		file:      file,
		cmdRunner: cmdRunner,
		done:      make(chan bool, 1),
	}

	err := fw.watch()
	return fw, err
}

// OnEvent ...
func (f FileWatcher) OnEvent() {
	log.Printf("change in file %v detected. Executing command...", f.file)
	stdout, stderr, err := f.cmdRunner.Run()
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
func (f FileWatcher) Close() error {
	f.done <- true
	return f.watcher.Close()
}

// watch creates a fsnotify watcher for a file and create of write events
func (f *FileWatcher) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	f.watcher = watcher

	dir, file := path.Split(f.file)
	go func(file string) {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create &&
						event.Name == file {
					f.OnEvent()
				}
			case err := <-watcher.Errors:
				if err != nil {
					log.Printf("error watching file: %v\n", err)
				}
			case <-f.done:
				log.Printf("file watcher done")
				return
			}
		}
	}(file)
	return watcher.Add(dir)
}
