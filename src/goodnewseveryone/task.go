package goodnewseveryone

import (
	"path/filepath"
	"os"
	"strings"
	"fmt"
	"io/ioutil"
	"errors"
	"encoding/json"
)

var (
	errDuplicateTask = errors.New("Duplicate Task")
	errUnknownTask = errors.New("Unknown Task")
)

type Tasks map[string]Task

func configToTasks(log Log, configLoc string) (Tasks, error) {
	tasks := make(Tasks)
	err := filepath.Walk(configLoc, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "task.json") {
			log.Write(fmt.Sprintf("Task Config: %v", path))
			task, err := configToTask(path)
			if err != nil {
				return err
			}
			log.Write(fmt.Sprintf("Task Configured: %v", task))
			if err := tasks.Add(task); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return tasks, nil
}

func (tasks Tasks) Remove(task Task) error {
	if _, ok := tasks[task.String()]; !ok {
		return errUnknownTask
	}
	delete(tasks, task.String())
	return nil
}

func (tasks Tasks) Add(task Task) error {
	if _, ok := tasks[task.String()]; ok {
		return errDuplicateTask
	}
	tasks[task.String()] = task
	return nil
}

type TaskType string

var (
	Sync = TaskType("sync")
	Backup = TaskType("backup")
)

var (
	errUndefinedTaskType = errors.New("Undefined Task Type: currently only sync and backup are supported")
)

type Task struct {
	Type TaskType
	Src string
	Dst string
}

func configToTask(filename string) (Task, error) {
	task := Task{}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return task, err
	}
	if err := json.Unmarshal(data, &task); err != nil {
		return task, err
	}
	if task.Type != Sync && task.Type != Backup {
		return task, errUndefinedTaskType
	}
	return task, nil
}

func NewTask(src string, typ TaskType, dst string) Task {
	return Task{typ, src, dst}
}

func (this Task) String() string {
	return fmt.Sprintf("%v --%v-> %v", this.Src, this.Type, this.Dst)
}

func (this Task) newCommand(locations Locations) (*command, error) {
	src, ok := locations[this.Src]
	if !ok {
		return nil, errUnknownLocation
	}
	dst, ok := locations[this.Dst]
	if !ok {
		return nil, errUnknownLocation
	}
	switch this.Type {
	case Sync:
		return newSyncCommand(src.getLocal(), dst.getLocal()), nil
	case Backup:
		return newBackupCommand(src.getLocal(), dst.getLocal()), nil
	}
	panic("unreachable")
}


