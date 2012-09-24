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
	errInvalidLocation = errors.New("Specified Location has not been configured")
)

type tasks []Task

func newTasks(log Log, locations locations, configLoc string) (tasks, error) {
	tasks := make(tasks, 0)
	err := filepath.Walk(configLoc, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "task.json") {
			log.Write(fmt.Sprintf("Task Config: %v", path))
			task, err := ConfigToTask(path)
			if err != nil {
				log.Error(err)
				return err
			}
			log.Write(fmt.Sprintf("Task Configured: %v", task))
			tasks = append(tasks, task)
			if _, ok := locations[task.Src]; !ok {
				log.Error(errInvalidLocation)
				return errInvalidLocation
			}
			if _, ok := locations[task.Dst]; !ok {
				log.Error(errInvalidLocation)
				return errInvalidLocation
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tasks, nil
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

func ConfigToTask(filename string) (Task, error) {
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

func (this Task) String() string {
	return fmt.Sprintf("%v --%v-> %v", this.Src, this.Type, this.Dst)
}

func (this Task) NewCommand() *command {
	switch this.Type {
	case Sync:
		return newSyncCommand(this.Src, this.Dst)
	case Backup:
		return newBackupCommand(this.Src, this.Dst)
	}
	panic("unreachable")
}


