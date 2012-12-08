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
	"testing"
	"goodnewseveryone/files"
	"goodnewseveryone/location"
	"goodnewseveryone/command"
	"goodnewseveryone/log"
	"time"
)

func TestLocations(t *testing.T) {
	f := files.NewFiles(".")
	gne := NewGNE(f)
	name := "a"
	loc := location.NewLocalLocation(name, ".")
	if err := gne.AddLocation(loc); err != nil {
		panic(err)
	}
	locations := gne.GetLocations()
	if len(locations) != 1 {
		t.Fatalf("Expected 1 location, but got %v", len(locations))
	}
	if l, ok := locations[name]; !ok {
		t.Fatalf("Expected %v, but does not exist in %v", name, locations)
	} else {
		if l.GetLocal() != "." {
			t.Fatalf("Expected local . , but got %v", l.GetLocal())
		}
	}
	if err := gne.RemoveLocation(name); err != nil {
		panic(err)
	}
	locations = gne.GetLocations()
	if len(locations) != 0 {
		t.Fatalf("Expected 0 location, but got %v", len(locations))
	}
}

type mockCommand struct {
	done chan bool
}

func (this *mockCommand) Run(log log.Log) ([]byte, error) {
	if this.done != nil {
		this.done <- true
	}
	return nil, nil
}

func (this *mockCommand) Stop(log log.Log) {
	return
}

type mockTask struct {
	name string
	cmd *mockCommand
	src string
	dst string
	completed time.Time
}

func (this *mockTask) Name() string {
	return this.name
}

func (this *mockTask) NewCommand(locations location.Locations) (command.Command, error) {
	return this.cmd, nil
}

func (this *mockTask) TaskTypeName() string {
	return "ataskTypeName"
}

func (this *mockTask) Src() string {
	return this.src
}

func (this *mockTask) Dst() string {
	return this.dst
}

func (this *mockTask) LastCompleted() time.Time {
	return this.completed
}

func (this *mockTask) Complete(completed time.Time) {
	this.completed = completed
}

func TestTasks(t *testing.T) {
	f := files.NewFiles(".")
	gne := NewGNE(f)
	name := "ataskname"
	cmd := &mockCommand{}
	src := "srcname"
	dst := "dstname"
	if err := gne.AddLocation(location.NewLocalLocation(src, ".")); err != nil {
		panic(err)
	}
	if err := gne.AddLocation(location.NewLocalLocation(dst, ".")); err != nil {
		panic(err)
	}
	task := &mockTask{
		name: name,
		cmd: cmd,
		src: src,
		dst: dst,
	}
	if err := gne.AddTask(task); err != nil {
		panic(err)
	}
	tasks := gne.GetTasks()
	taskNames := tasks.List()
	if len(taskNames) != 1 {
		t.Fatalf("Expected 1, but got %v tasks", len(taskNames))
	}
	if taskNames[0] != name {
		t.Fatalf("Expedted %v, but got %v", name, taskNames[0])
	}
	if err := gne.RemoveTask(name); err != nil {
		panic(err)
	}
	tasks = gne.GetTasks()
	taskNames = tasks.List()
	if len(taskNames) != 0 {
		t.Fatalf("Expected 0, but got %v tasks", len(taskNames))
	} 
	if err := gne.RemoveLocation(src); err != nil {
		panic(err)
	}
	if err := gne.RemoveLocation(dst); err != nil {
		panic(err)
	}
}

func TestDiffs(t *testing.T) {
	f := files.NewFiles(".")
	gne := NewGNE(f)
	go gne.Start()
	name := "ataskname"
	done := make(chan bool)
	cmd := &mockCommand{done}
	src := "srcname"
	dst := "dstname"
	if err := gne.AddLocation(location.NewLocalLocation(src, ".")); err != nil {
		panic(err)
	}
	if err := gne.AddLocation(location.NewLocalLocation(dst, ".")); err != nil {
		panic(err)
	}
	task := &mockTask{
		name: name,
		cmd: cmd,
		src: src,
		dst: dst,
	}
	if err := gne.AddTask(task); err != nil {
		panic(err)
	}
	go gne.Now(name)
	<- done
	go gne.Now(name)
	<- done
	diffs, err := gne.GetDiffs()
	if err != nil {
		panic(err)
	}
	if len(diffs) == 0 {
		t.Fatalf("diffs = %v", diffs)
	}

	if err := gne.RemoveTask(name); err != nil {
		panic(err)
	}
	if err := gne.RemoveLocation(src); err != nil {
		panic(err)
	}
	if err := gne.RemoveLocation(dst); err != nil {
		panic(err)
	}
}

func TestLogs(t *testing.T) {
	f := files.NewFiles(".")
	gne := NewGNE(f)
	logs, err := gne.GetLogs()
	if err != nil {
		panic(err)
	}
	for _, l := range logs {
		if _, err := l.Open(); err != nil {
			panic(err)
		}
	}
}

func TestExecution(t *testing.T) {
	/*SetWaitTime(waitTime time.Duration)
	GetWaitTime() time.Duration

	Now(taskName string)
	Unblock()
	StopAndBlock()
	Blocked() bool
	BusyWith() (taskName string)

	Start()*/
	panic("todo")
}
