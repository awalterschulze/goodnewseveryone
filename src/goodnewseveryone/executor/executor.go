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

package executor

import (
	"sync"
	"fmt"
	"goodnewseveryone/kernel"
	"goodnewseveryone/log"
	"goodnewseveryone/task"
	"goodnewseveryone/location"
	"goodnewseveryone/diff"
	gstore "goodnewseveryone/store"
	"time"
)

type Executor interface {
	Blocked() bool
	StopAndBlock(log log.Log)
	Unblock()
	BusyWith() (taskName string)
	Execute(log log.Log, task task.Task, locations location.Locations, store gstore.FilelistStore)
}

type executor struct {
	sync.Mutex
	kernel kernel.Kernel
	busy string
}

func NewExecutor(kernel kernel.Kernel) *executor {
	return &executor{
		kernel: kernel,
	}
}

func (this *executor) Blocked() bool {
	if err := this.kernel.Blocked(); err != nil {
		return true
	}
	return false
}

func (this *executor) StopAndBlock(log log.Log) {
	this.kernel.StopAndBlock(log)
	//wait until stopped task returns from execute method and this.busy is set to nil
	this.Lock()
	this.Unlock()
}
	
func (this *executor) Unblock() {
	this.kernel.Unblock()
}

func (this *executor) BusyWith() string {
	return this.busy
}

func (this *executor) Execute(log log.Log, task task.Task, locations location.Locations, store gstore.FilelistStore) {
	this.Lock()
	log.Write(fmt.Sprintf("Executing Task %v", task))
	this.busy = task.Name()
	if err := this.execute(log, task, locations, store); err != nil {
		log.Error(err)
	}
	this.busy = ""
	log.Write(fmt.Sprintf("Executed Task %v", task))
	this.Unlock()
	return
}

func (this *executor) execute(log log.Log, task task.Task, locations location.Locations, store gstore.FilelistStore) error {
	if err := this.kernel.Blocked(); err != nil {
		return err
	}
	src, ok := locations[task.Src()]
	if !ok {
		log.Error(gstore.ErrLocationDoesNotExist)
		return gstore.ErrLocationDoesNotExist
	}
	dst, ok := locations[task.Dst()]
	if !ok {
		log.Error(gstore.ErrLocationDoesNotExist)
		return gstore.ErrLocationDoesNotExist
	}
	output, err := this.kernel.Run(log, src.NewPreparedCommand())
	if err != nil || !src.Prepared(log, output) {
		if _, err := this.kernel.Run(log, src.NewPrepareCommand()); err != nil {
			return err
		}
	}
	output, err = this.kernel.Run(log, dst.NewPreparedCommand())
	if err != nil || !dst.Prepared(log, output) {
		if _, err := this.kernel.Run(log, dst.NewPrepareCommand()); err != nil {
			return err
		}
	}
	this.kernel.SudoRun(log, src.NewUmountCommand())
	this.kernel.SudoRun(log, dst.NewUmountCommand())
	output, err = this.kernel.Run(log, src.NewLocatedCommand())
	if err != nil {
		return err
	}
	if !src.Located(log, output) {
		return nil
	}
	output, err = this.kernel.Run(log, dst.NewLocatedCommand())
	if err != nil {
		return err
	}
	if !dst.Located(log, output) {
		return nil
	}
	_, err = this.kernel.Run(log, src.NewMountCommand())
	if err != nil {
		return err
	}
	defer this.kernel.SudoRun(log, src.NewUmountCommand())
	_, err = this.kernel.Run(log, dst.NewMountCommand())
	if err != nil {
		return err
	}
	defer this.kernel.SudoRun(log, dst.NewUmountCommand())
	taskCommand, err := task.NewCommand(locations)
	if err != nil {
		return err
	}
	srcList, err := diff.CreateFilelist(src.GetLocal())
	if err != nil {
		return err
	}
	if err := diff.SaveFilelist(store, string(src.Id()), time.Now(), srcList); err != nil {
		return err
	}
	dstList, err := diff.CreateFilelist(dst.GetLocal())
	if err != nil {
		return err
	}
	if err := diff.SaveFilelist(store, string(dst.Id()), time.Now(), dstList); err != nil {
		return err
	}
	if _, err := this.kernel.Run(log, taskCommand); err != nil {
		return err
	}
	srcList, err = diff.CreateFilelist(src.GetLocal())
	if err != nil {
		return err
	}
	if err := diff.SaveFilelist(store, string(src.Id()), time.Now(), srcList); err != nil {
		return err
	}
	dstList, err = diff.CreateFilelist(dst.GetLocal())
	if err != nil {
		return err
	}
	if err := diff.SaveFilelist(store, string(dst.Id()), time.Now(), dstList); err != nil {
		return err
	}
	task.Complete(time.Now())
	return nil
}
