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

type TaskId string

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

type Tasks struct {
	store gstore.TaskStore
	tasks map[TaskId]Task
}

func NewTasks(store gstore.TaskStore) (Tasks, error) {
	taskTypes, err := ListTaskTypes(store)
	if err != nil {
		return Tasks{}, err
	}
	tasks := make(map[TaskId]Task)
	taskIds, err := store.ListTasks()	
	if err != nil {
		return Tasks{}, err
	}
	for _, taskId := range taskIds {
		src, taskTypName, dst, err := store.ReadTask(taskId)
		if err != nil {
			return Tasks{}, err
		}
		var taskType TaskType = nil
		for _, taskTyp := range taskTypes {
			if taskTypName == taskTyp.Name() {
				taskType = taskTyp
			}
		}
		if taskType.Name() != taskTypName {
			return Tasks{}, gstore.ErrTaskTypeDoesNotExist
		}
		task := &task{
			name: TaskId(taskId),
			src: location.LocationId(src),
			typ: taskType,
			dst: location.LocationId(dst),
		}
		times, err := store.ListTaskCompleted(taskId)
		if err != nil {
			return Tasks{}, err
		}
		for i, t := range times {
			if t.After(task.lastCompleted) {
				task.lastCompleted = times[i]
			}
		}
		tasks[TaskId(taskId)] = task
	}
	return Tasks{
		store: store,
		tasks: tasks,
	}, nil
}

func (tasks Tasks) Add(task Task) error {
	if _, ok := tasks.tasks[task.Id()]; ok {
		return gstore.ErrTaskAlreadyExists
	}
	if err := tasks.store.AddTask(string(task.Id()), string(task.Src()), task.TaskTypeName(), string(task.Dst())); err != nil {
		return err
	}
	tasks.tasks[task.Id()] = task
	return nil
}

func (tasks Tasks) Remove(taskId TaskId) error {
	if _, ok := tasks.tasks[taskId]; !ok {
		return gstore.ErrTaskDoesNotExist
	}
	times, err := tasks.store.ListTaskCompleted(string(taskId))
	if err != nil {
		return err
	}
	for _, t := range times {
		if err := tasks.store.RemoveTaskCompleted(string(taskId), t); err != nil {
			return err
		}
	}
	if err := tasks.store.RemoveTask(string(taskId)); err != nil {
		return err
	}
	delete(tasks.tasks, taskId)
	return nil
}

func (tasks Tasks) Complete(taskId TaskId, now time.Time) error {
	if _, ok := tasks.tasks[taskId]; !ok {
		return gstore.ErrTaskDoesNotExist
	}
	if err := tasks.store.AddTaskCompleted(string(taskId), now); err != nil {
		return err
	}
	t := tasks.tasks[taskId]
	if now.After(t.LastCompleted()) {
		t.Complete(now)
	}
	tasks.tasks[taskId] = t
	return nil
}

type Task interface {
	Id() TaskId
	NewCommand(locations location.Locations) (command.Command, error)
	TaskTypeName() string
	Src() location.LocationId
	Dst() location.LocationId
	LastCompleted() time.Time
	Complete(time.Time)
}

type task struct {
	name TaskId
	typ TaskType
	src location.LocationId
	dst location.LocationId
	lastCompleted time.Time
}

func NewTask(name TaskId, typ TaskType, src, dst location.LocationId) Task {
	return &task{
		name: name,
		typ: typ,
		src: src,
		dst: dst,
	}
}

func (this *task) Id() TaskId {
	return TaskId(this.name)
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

func (this *task) Src() location.LocationId {
	return this.src
}

func (this *task) Dst() location.LocationId {
	return this.dst
}

func (this *task) LastCompleted() time.Time {
	return this.lastCompleted
}
	
func (this *task) Complete(now time.Time) {
	this.lastCompleted = now
}


