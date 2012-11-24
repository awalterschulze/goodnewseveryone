//Copyright 2012 Walter Schulze
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package goodnewseveryone

import (
	"sync"
	"errors"
)

var (
	errPaused = errors.New("Paused")
)

type kernel struct {
	sync.Mutex
	running *command
	paused bool
}

func newKernel() *kernel {
	return &kernel{}
}

func (this *kernel) ready() bool {
	this.Lock()
	defer this.Unlock()
	return !this.paused
}

func (this *kernel) restart() {
	this.Lock()
	this.paused = false
	this.Unlock()
}

func (this *kernel) stop(log Log) {
	this.Lock()
	this.paused = true
	if this.running != nil {
		this.running.stop(log)
	}
	this.Unlock()
}

func (this *kernel) run(log Log, command *command) (string, error) {
	this.Lock()
	if this.paused {
		this.Unlock()
		return "", errPaused
	}
	this.running = command
	this.Unlock()
	if this.running == nil {
		return "", nil
	}
	output, err := this.running.run(log)
	if err != nil {
		return "", err
	}
	this.Lock()
	this.running = nil
	this.Unlock()
	return string(output), err
}

func (this *kernel) overrun(log Log, command *command) {
	this.Lock()
	defer this.Unlock()
	if command == nil {
		return
	}
	this.running = command
	this.running.run(log)
	this.running = nil
}
