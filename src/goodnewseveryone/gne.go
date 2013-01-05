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
	"goodnewseveryone/diff"
	"goodnewseveryone/executor"
	"goodnewseveryone/kernel"
	"goodnewseveryone/location"
	"goodnewseveryone/log"
	gstore "goodnewseveryone/store"
	"goodnewseveryone/task"
	"sync"
	"time"
)

type GNE interface {
	AddLocation(loc location.Location) error
	RemoveLocation(locName string) error
	GetLocations() location.Locations
	GetRemoteLocationTypes() ([]location.RemoteLocationType, error)

	AddTask(task task.Task) error
	RemoveTask(taskName string) error
	GetTasks() task.Tasks

	GetTaskTypes() (types []task.TaskType, err error)

	SetWaitTime(waitTime time.Duration) error
	GetWaitTime() time.Duration

	Now(taskName string)
	Unblock()
	StopAndBlock()
	Blocked() bool
	BusyWith() (taskName string)

	GetLogs() (log.LogContents, error)

	GetDiffs() (diff.DiffsPerLocation, error)

	GetMountFolder() (string, error)

	Start()
	End()
}

type gne struct {
	sync.Mutex
	store     gstore.Store
	locations location.Locations
	tasks     task.Tasks
	executor  executor.Executor
	waitTime  time.Duration
	waitChan  <-chan time.Time
	nowChan   chan string
	endChan   chan bool
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
		store:     store,
		locations: locations,
		tasks:     tasks,
		executor:  executor.NewExecutor(kernel.NewKernel()),
		waitTime:  waitTime,
		waitChan:  time.After(waitTime),
		nowChan:   make(chan string),
		endChan:   make(chan bool),
	}
	return gne
}

func (this *gne) GetMountFolder() (string, error) {
	return this.store.GetMountFolder()
}

func (this *gne) AddLocation(loc location.Location) error {
	return this.locations.Add(this.store, loc)
}

func (this *gne) RemoveLocation(locId string) error {
	return this.locations.Remove(this.store, locId)
}

func (this *gne) GetLocations() location.Locations {
	return this.locations
}

func (this *gne) GetRemoteLocationTypes() ([]location.RemoteLocationType, error) {
	return location.ListRemoteLocationTypes(this.store)
}

func (this *gne) AddTask(task task.Task) error {
	if _, ok := this.locations[task.Src()]; !ok {
		return gstore.ErrLocationDoesNotExist
	}
	if _, ok := this.locations[task.Dst()]; !ok {
		return gstore.ErrLocationDoesNotExist
	}
	return this.tasks.Add(task)
}

func (this *gne) RemoveTask(taskName string) error {
	return this.tasks.Remove(taskName)
}

func (this *gne) GetTasks() task.Tasks {
	return this.tasks
}

func (this *gne) GetTaskTypes() (types []task.TaskType, err error) {
	return task.ListTaskTypes(this.store)
}

func (this *gne) SetWaitTime(waitTime time.Duration) error {
	err := this.store.SetWaitTime(waitTime)
	if err != nil {
		return err
	}
	this.waitTime = waitTime
	this.waitChan = time.After(this.waitTime)
	return nil
}

func (this *gne) GetWaitTime() time.Duration {
	return this.waitTime
}

func (this *gne) Now(taskName string) {
	this.nowChan <- taskName
}

func (this *gne) Unblock() {
	this.executor.Unblock()
}

func (this *gne) StopAndBlock() {
	l, _ := log.NewLog(time.Now(), this.store)
	this.executor.StopAndBlock(l)
}

func (this *gne) Blocked() bool {
	return this.executor.Blocked()
}

func (this *gne) BusyWith() string {
	return this.executor.BusyWith()
}

func (this *gne) GetLogs() (log.LogContents, error) {
	return log.NewLogContents(this.store)
}

func (this *gne) GetDiffs() (diff.DiffsPerLocation, error) {
	return diff.NewDiffsPerLocation(this.store)
}

func (this *gne) runAll() {
	tasks := this.tasks.List()
	for _, taskName := range tasks {
		this.run(taskName)
	}
}

func (this *gne) run(taskName string) {
	l, err := log.NewLog(time.Now(), this.store)
	if err != nil {
		panic(err)
	}
	task := this.tasks.Get(taskName)
	if task == nil {
		l.Write(fmt.Sprintf("Task %v Does not Exist", task))
		return
	}
	this.executor.Execute(l, task, this.locations, this.store)
}

func (this *gne) Start() {
	for {
		select {
		case <-this.waitChan:
			this.runAll()
			this.waitChan = time.After(this.GetWaitTime())
		case t := <-this.nowChan:
			this.run(t)
		case <-this.endChan:
			return
		}

	}
}

func (this *gne) End() {
	this.endChan <- true
}
