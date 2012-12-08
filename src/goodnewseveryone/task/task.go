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

package task

import (
	"time"
	gstore "goodnewseveryone/store"
	"goodnewseveryone/location"
	"goodnewseveryone/command"
)

type TaskType interface {
	NewCommand(src, dst string) command.Command
	Name() string
	CmdStr() string
}

type taskType struct {
	name string
	cmdStr string
}

func (this *taskType) NewCommand(src, dst string) command.Command {
	return command.NewCommand(this.cmdStr, src, dst)
}

func (this *taskType) Name() string {
	return this.name
}

func (this *taskType) CmdStr() string {
	return this.cmdStr
}

func ListTaskTypes(store gstore.TaskStore) (types []TaskType, err error) {
	names, err := store.ListTaskTypes()
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		cmdStr, err := store.ReadTaskType(name)
		if err != nil {
			return nil, err
		}
		types = append(types, &taskType{
			name: name,
			cmdStr: cmdStr,
		})
	}
	return types, nil
}

func AddTaskType(store gstore.TaskStore, taskType TaskType) error {
	return store.AddTaskType(taskType.Name(), taskType.CmdStr())
}

func RemoveTaskType(store gstore.TaskStore, name string) error {
	return store.RemoveTaskType(name)
}

type Tasks interface {
	List() []string
	Get(taskName string) Task
	Add(task Task) error
	Remove(taskName string) error
	Complete(taskName string, now time.Time) error
}

type tasks struct {
	store gstore.TaskStore
	tasks map[string]Task
}

func NewTasks(store gstore.TaskStore) (*tasks, error) {
	taskTypes, err := ListTaskTypes(store)
	if err != nil {
		return nil, err
	}
	ts := make(map[string]Task)
	taskNames, err := store.ListTasks()	
	if err != nil {
		return nil, err
	}
	for _, taskName := range taskNames {
		src, taskTypName, dst, err := store.ReadTask(taskName)
		if err != nil {
			return nil, err
		}
		var taskType TaskType = nil
		for _, taskTyp := range taskTypes {
			if taskTypName == taskTyp.Name() {
				taskType = taskTyp
			}
		}
		if taskType.Name() != taskTypName {
			return nil, gstore.ErrTaskTypeDoesNotExist
		}
		task := &task{
			name: taskName,
			src: src,
			typ: taskType,
			dst: dst,
		}
		times, err := store.ListTaskCompleted(taskName)
		if err != nil {
			return nil, err
		}
		for i, t := range times {
			if t.After(task.lastCompleted) {
				task.lastCompleted = times[i]
			}
		}
		ts[taskName] = task
	}
	return &tasks{
		store: store,
		tasks: ts,
	}, nil
}

func (tasks *tasks) Get(taskName string) Task {
	return tasks.tasks[taskName]
}

func (tasks *tasks) List() []string {
	list := make([]string, 0, len(tasks.tasks))
	for id, _ := range tasks.tasks {
		list = append(list, id)
	}
	return list
}

func (tasks *tasks) Add(task Task) error {
	if _, ok := tasks.tasks[task.Name()]; ok {
		return gstore.ErrTaskAlreadyExists
	}
	if err := tasks.store.AddTask(string(task.Name()), string(task.Src()), task.TaskTypeName(), string(task.Dst())); err != nil {
		return err
	}
	tasks.tasks[task.Name()] = task
	return nil
}

func (tasks *tasks) Remove(taskName string) error {
	if _, ok := tasks.tasks[taskName]; !ok {
		return gstore.ErrTaskDoesNotExist
	}
	times, err := tasks.store.ListTaskCompleted(string(taskName))
	if err != nil {
		return err
	}
	for _, t := range times {
		if err := tasks.store.RemoveTaskCompleted(string(taskName), t); err != nil {
			return err
		}
	}
	if err := tasks.store.RemoveTask(string(taskName)); err != nil {
		return err
	}
	delete(tasks.tasks, taskName)
	return nil
}

func (tasks *tasks) Complete(taskName string, now time.Time) error {
	if _, ok := tasks.tasks[taskName]; !ok {
		return gstore.ErrTaskDoesNotExist
	}
	if err := tasks.store.AddTaskCompleted(string(taskName), now); err != nil {
		return err
	}
	t := tasks.tasks[taskName]
	if now.After(t.LastCompleted()) {
		t.Complete(now)
	}
	tasks.tasks[taskName] = t
	return nil
}

type Task interface {
	Name() string
	NewCommand(locations location.Locations) (command.Command, error)
	TaskTypeName() string
	Src() string
	Dst() string
	LastCompleted() time.Time
	Complete(time.Time)
}

type task struct {
	name string
	typ TaskType
	src string
	dst string
	lastCompleted time.Time
}

func NewTask(name string, typ TaskType, src, dst string) Task {
	return &task{
		name: name,
		typ: typ,
		src: src,
		dst: dst,
	}
}

func (this *task) Name() string {
	return this.name
}

func (this *task) NewCommand(locations location.Locations) (command.Command, error) {
	src, ok := locations[this.src]
	if !ok {
		return nil, gstore.ErrLocationDoesNotExist
	}
	dst, ok := locations[this.dst]
	if !ok {
		return nil, gstore.ErrLocationDoesNotExist
	}
	return this.typ.NewCommand(src.GetLocal(), dst.GetLocal()), nil
}

func (this *task) TaskTypeName() string {
	return this.typ.Name()
}

func (this *task) Src() string {
	return this.src
}

func (this *task) Dst() string {
	return this.dst
}

func (this *task) LastCompleted() time.Time {
	return this.lastCompleted
}
	
func (this *task) Complete(now time.Time) {
	this.lastCompleted = now
}


