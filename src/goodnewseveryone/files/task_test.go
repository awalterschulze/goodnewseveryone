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

package files

import (
	"testing"
	"time"
)

func TestTaskTypes(t *testing.T) {
	f := NewFiles(".")
	name := "move"
	taskType := "rsync -r --remove-source-files %v %v"
	if err := f.AddTaskType(name, taskType); err != nil {
		panic(err)
	}
	names, err := f.ListTaskTypes()
	if err != nil {
		panic(err)
	}
	if len(names) != 1 {
		t.Fatalf("wrong number of taskTypes")
	}
	if names[0] != name {
		t.Fatalf("not the correct name, expected %v, but got %v", name, names[0])
	}
	tt, err := f.ReadTaskType(name)
	if err != nil {
		panic(err)
	}
	if tt != taskType {
		t.Fatalf("wrong task Type expected %v, but got %v", taskType, tt)
	}
	if err := f.RemoveTaskType(name); err != nil {
		panic(err)
	}
	names, err = f.ListTaskTypes()
	if err != nil {
		panic(err)
	}
	if len(names) != 0 {
		t.Fatalf("task was not deleted")
	}
}

func TestTasks(t *testing.T) {
	f := NewFiles(".")
	taskName := "a"
	src := "/home/b"
	dst := "/home/c/"
	tt := "move"
	if err := f.AddTask(taskName, src, tt, dst); err != nil {
		panic(err)
	}
	ids, err := f.ListTasks()
	if err != nil {
		panic(err)
	}
	if len(ids) != 1 {
		t.Fatalf("wrong number of tasks")
	}
	if ids[0] != taskName {
		t.Fatalf("not the correct id, expected %v, but got %v", taskName, ids[0])
	}
	src1, tt1, dst1, err := f.ReadTask(taskName)
	if err != nil {
		panic(err)
	}
	if src1 != src || dst1 != dst || tt1 != tt {
		t.Fatalf("wrong task %v %v %v != %v %v %v", src, dst, tt, src1, dst1, tt1)
	}
	if err := f.RemoveTask(taskName); err != nil {
		panic(err)
	}
	ids, err = f.ListTasks()
	if err != nil {
		panic(err)
	}
	if len(ids) != 0 {
		t.Fatalf("task was not deleted")
	}
}

func TestCompleted(t *testing.T) {
	f := NewFiles(".")
	task := "a"
	now := time.Now()
	if err := f.AddTaskCompleted(task, now); err != nil {
		panic(err)
	}
	ts, err := f.ListTaskCompleted(task)
	if err != nil {
		panic(err)
	}
	if len(ts) != 1 {
		t.Fatalf("completed task not found = %v", ts)
	}
	if err := f.RemoveTaskCompleted(task, ts[0]); err != nil {
		panic(err)
	}
	ts, err = f.ListTaskCompleted(task)
	if err != nil {
		panic(err)
	}
	if len(ts) != 0 {
		t.Fatalf("completed task not removed")
	}
}
