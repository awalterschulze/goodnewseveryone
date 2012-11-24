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
	"fmt"
	"time"
	"goodnewseveryone/store"
	"goodnewseveryone/location"
	"goodnewseveryone/command"
)

type TaskId string

type TaskType struct {
	Name string
	CmdStr string
}

func (this TaskType) NewCommand(src, dst string) command.Command {
	return fmt.Sprintf(this.CmdStr, src, dst)
}

func ListTaskTypes(store store.TaskStore) (types []TaskType, err error) {
	names, err := store.ListTaskTypes()
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		cmdStr, err := store.ReadTaskType(name)
		if err != nil {
			return nil, err
		}
		types = append(types, TaskType{
			Name: name,
			CmdStr: cmdStr,
		})
	}
	return types, nil
}

func AddTaskType(store store.TaskStore, taskType TaskType) error {
	return store.AddTaskType(taskType.Name, taskType.CmdStr)
}

func RemoveTaskType(store store.TaskStore, name string) error {
	return store.RemoveTaskType(name)
}

type Tasks struct {
	store store.TaskStore
	tasks map[TaskId]Task
}

type NewTasks(store store.TaskStore) (Tasks, error) {
	tasks := make(Tasks)
	taskIds, err := store.ListTasks()	
	if err != nil {
		return nil, err
	}
	for _, taskId := range taskIds {
		src, typ, dst, err := store.ReadTask(taskId)
		if err != nil {
			return nil, err
		}
		task := Task{
			Src: src,
			Typ: typ,
			Dst: dst,
		}
		times, err := store.ListTaskCompleted(taskId)
		if err != nil {
			return nil, err
		}
		for i, t := range times {
			if t.After(task.LastCompleted) {
				task.LastCompleted = times[i]
			}
		}
		tasks[taskId] = task
	}
	return Tasks{
		store: store,
		tasks: tasks,
	}, nil
}

var (
	errDuplicateTask = errors.New("Duplicate Task")
)

func (tasks Tasks) Add(task Task) error {
	if _, ok := tasks[task.Id()]; ok {
		return errDuplicateTask
	}
	if err := tasks.store.AddTask(task.Src, task.Typ.Name, task.Dst); err != nil {
		return err
	}
	tasks[task.Id()] = task
	return nil
}

var (
	errUnknownTask = errors.New("Unknown Task")
)

func (tasks Tasks) Remove(taskId TaskId) error {
	if _, ok := tasks[taskId]; !ok {
		return errUnknownTask
	}
	if err := tasks.store.RemoveTask(taskId); err != nil {
		return err
	}
	delete(tasks.tasks, taskId)
	return nil
}

func (tasks Tasks) Complete(taskId TaskId) error {
	return tasks.store.AddTaskCompleted(taskId, time.Now())
}

type Task struct {
	Typ TaskType
	Src location.LocationId
	Dst location.LocationId
	LastCompleted time.Time
}

func (this Task) Id() TaskId {
	return TaskId(fmt.Sprintf("%v---%v---%v", this.Src, this.Typ.Name, this.Dst))	
}

var (
	errUnknownLocation = errors.New("Unknown Location")
)

func (this Task) NewCommand(locations location.Locations) (command.Command, error) {
	src, ok := locations[this.Src]
	if !ok {
		return nil, errUnknownLocation
	}
	dst, ok := locations[this.Dst]
	if !ok {
		return nil, errUnknownLocation
	}
	return this.Typ.NewCommand(src.GetLocal(), dst.GetLocal())
}

