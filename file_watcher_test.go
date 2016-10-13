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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

type mockCmdRunner struct {
}

func newMockRunner() CmdRunner {
	return mockCmdRunner{}
}

func (mr mockCmdRunner) Run() (string, string, error) {
	log.Printf("run")
	return "", "", nil
}

func TestFileWatcher(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data := []struct {
		file   string
		change bool
	}{
		{fmt.Sprintf("%v/test-file-%v", dir, time.Now().Unix()), false},
		{fmt.Sprintf("%v/test-file-%v", dir, time.Now().Unix()), true},
	}

	for _, test := range data {
		done := make(chan bool)
		mr := newMockRunner()

		fw, err := NewFileWatcher(test.file, mr)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		eData := ""
		if test.change {
			eData = "dummy"
			go func() {
				err := ioutil.WriteFile(test.file, []byte("dummy"), 0777)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				done <- true
			}()

			<-done

			data, err := ioutil.ReadFile(test.file)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual([]byte(eData), data) {
				t.Error("expected to be equals")
			}

			os.Remove(test.file)
		}

		fw.Close()
	}
}
