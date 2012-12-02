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

package kernel

import (
	"sync"
	"errors"
	"goodnewseveryone/command"
	"goodnewseveryone/log"
)

var (
	ErrPaused = errors.New("Kernel is Paused")
)

type Kernel interface {
	Blocked() bool
	StopAndBlock(log log.Log)
	Unblock()
	Run(log log.Log, command command.Command) (string, error)
	//run even if blocked, used for umounts
	SudoRun(log log.Log, command command.Command)
}

type kernel struct {
	sync.Mutex
	running command.Command
	blocked bool
}

func NewKernel() Kernel {
	return &kernel{}
}

func (this *kernel) Blocked() bool {
	this.Lock()
	defer this.Unlock()
	return this.blocked
}

func (this *kernel) Unblock() {
	this.Lock()
	this.blocked = false
	this.Unlock()
}

func (this *kernel) StopAndBlock(log log.Log) {
	this.Lock()
	this.blocked = true
	if this.running != nil {
		this.running.Stop(log)
	}
	this.running = nil
	this.Unlock()
}

func (this *kernel) Run(log log.Log, command command.Command) (string, error) {
	this.Lock()
	if this.blocked {
		defer this.Unlock()
		return "", ErrPaused
	}
	this.running = command
	this.Unlock()
	if this.running == nil {
		return "", nil
	}
	output, err := this.running.Run(log)
	if err != nil {
		return "", err
	}
	this.Lock()
	this.running = nil
	this.Unlock()
	return string(output), err
}

func (this *kernel) SudoRun(log log.Log, command command.Command) {
	this.Lock()
	defer this.Unlock()
	if command == nil {
		return
	}
	this.running = command
	this.running.Run(log)
	this.running = nil
}
