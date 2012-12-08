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
	"testing"
	"goodnewseveryone/files"
	"goodnewseveryone/log"
	"goodnewseveryone/task"
	"goodnewseveryone/location"
	"goodnewseveryone/command"
	"goodnewseveryone/kernel"
	"fmt"
	"time"
)

var (
	wait chan bool = make(chan bool)
)

type cmd struct {

}

func newCommand() *cmd {
	return &cmd{}
}

func (this *cmd) Run(log log.Log) ([]byte, error) {
	<-wait
	return nil, nil
}

func (this *cmd) Stop(log log.Log) {
	wait <- true
}

type taskType struct {
	name string
}

func (this *taskType) NewCommand(src, dst string) command.Command {
	return newCommand()
}
	
func (this *taskType) Name() string {
	return this.name
}
	
func (this *taskType) CmdStr() string {
	panic("not implemented")
}

type loggy struct {

}

func (this *loggy) Write(str string) {
	fmt.Printf(str+"\n")
}

func (this *loggy) Run(name string, arg ...string) {
	panic(name)
}

func (this *loggy) Error(err error) {
	if err == kernel.ErrBlocked {
		return
	}
	panic(err)
}

func (this *loggy) Output(output []byte) {
	this.Write(string(output))
}

func (this *loggy) Close() {

}

func TestExecutor(t *testing.T) {
	f := files.NewFiles(".")
	k := kernel.NewKernel()
	e := NewExecutor(k)
	id := "taskname"
	task := task.NewTask(id, &taskType{"typename"}, "a", "b")
	l := &loggy{}
	loc1 := location.NewLocalLocation("a", ".")
	loc2 := location.NewLocalLocation("b", ".")
	locations := make(location.Locations)
	locations[loc1.Id()] = loc1
	locations[loc2.Id()] = loc2
	go func() {
		e.Execute(l, task, locations, f)
	}()
	busy := e.BusyWith()
	for {
		if len(busy) > 0 {
			break
		}
		time.Sleep(1e6)
		busy = e.BusyWith()
	}
	if string(busy) != string(id) {
		t.Fatalf("Expected %v, but got %v", id, busy)
	}
	e.StopAndBlock(l)
	if len(e.BusyWith()) > 0 {
		t.Fatalf("Task was stopped")
	}
	if !e.Blocked() {
		t.Fatalf("Expected Blocked")
	}
	e.Unblock()
	w2 := make(chan bool)
	go func() {
		e.Execute(l, task, locations, f)
		close(w2)
	}()
	wait <- true
	<- w2
	if len(e.BusyWith()) > 0 {
		t.Fatalf("Expected done")
	}
}
