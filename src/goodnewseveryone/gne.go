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
	"fmt"
	"time"
	"sync"
	"errors"
	gstore "goodnewseveryone/store"
	"goodnewseveryone/log"
	"goodnewseveryone/location"
	"goodnewseveryone/kernel"
	"goodnewseveryone/task"
	"goodnewseveryone/diff"
	"goodnewseveryone/executor"
)

type GNE interface {
	AddLocation(loc Location) error
	RemoveLocation(locId LocationId) error
	GetLocations() Locations
	AddTask(task Task) error
	RemoveTask(taskId TaskId) error
	GetTasks() Tasks
	SetWaitTime(waitTime time.Duration)
	GetWaitTime() time.Duration
	Now(taskId TaskId)
	Unblock()
	Block()
	Blocked() bool
	BusyWith() TaskId
	GetLogs() (LogFiles, error)
	GetDiffs() (DiffsPerLocation, error)
	Start()
}

type gne struct {
	sync.Mutex
	locations location.Locations
	tasks task.Tasks
	executor executor.Executor
	waitTime time.Duration
	nowChan chan time.Time
	stopChan chan time.Time
	restartChan chan time.Time
}

func NewGNE(store gstore.Store) GNE {
	startupLog, err := log.NewLog(time.Now(), store)
	if err != nil {
		panic(err)
	}
	locations, err := location.NewLocations(startupLog, store)
	if err != nil {
		panic(err)
	}
	waitTime, err := store.GetWaitTime()
	if err != nil {
		panic(err)
	}
	tasks, err := task.NewTasks(store)
	if err != nil {
		panic(err)
	}

	gne := &gne{
		locations: locations,
		tasks: tasks,
		executor: executor.newExecutor(kernel.NewKernel()),
		waitTime: waitTime,
		nowChan: make(chan time.Time),
		stopChan: make(chan time.Time),
		restartChan: make(chan time.Time),
	}
}

func (this *gne) AddLocation(loc Location) error {
	return this.locations.Add(loc)
}

func (this *gne) RemoveLocation(locId LocationId) error {
	for _, t := range this.tasks {
		if t.Src == locId || t.Dst == locId {
			return errLocationExists
		}
	}
	return this.locations.Remove(locId)
}

func (this *gne) GetLocations() Locations {
	return this.locations
}

func (this *gne) AddTask(task Task) error {
	if _, ok := this.locations[task.Src]; !ok {
		return errUnknownLocation
	}
	if _, ok := this.locations[task.Dst]; !ok {
		return errUnknownLocation
	}
	return this.tasks.Add(task)
}

func (this *gne) RemoveTask(taskId TaskId) error {
	return this.tasks.Remove(taskId)
}

func (this *gne) GetTasks() Tasks {
	return this.tasks
}

func (this *gne) SetWaitTime(waitTime time.Duration) {
	this.waitTime = waitTime
}

func (this *gne) GetWaitTime() time.Duration {
	return this.waitTime
}

func (this *gne) Now() {
	this.nowChan <- time.Now()
}

func (this *gne) Restart() {
	this.restartChan <- time.Now()
}

func (this *gne) Stop() {
	this.stopChan <- time.Now()
}

func (this *gne) IsReady() bool {
	return this.kernel.ready()
}

func (this *gne) IsRunning() bool {
	return this.executor.IsRunning()
}

func (this *gne) GetLogs() (LogFiles, error) {
	return NewLogFiles(".")
}

func (this *gne) GetFileLists() (FileLists, error) {
	return NewFileLists(".")
}

func (this *gne) GetDiffs() (DiffsPerLocation, error) {
	return NewDiffsPerLocation(".")
}

func (this *gne) Start() {
	waitChan := time.After(1)
	//TODO executor.All should run in a go routine
	for {
		select {
		case <- waitChan:
			if this.kernel.ready() {
				this.executor.All(this.kernel, this.locations, this.tasks)
			}
		case <- this.nowChan:
			this.executor.All(this.kernel, this.locations, this.tasks)
		case <- this.stopChan:
			stopLog, err := newLog()
			if err != nil {
				panic(err)
			}
			this.kernel.stop(stopLog)
		case <- this.restartChan:
			this.kernel.restart()
		}
		waitChan = time.After(this.waitTime)
	}
}
