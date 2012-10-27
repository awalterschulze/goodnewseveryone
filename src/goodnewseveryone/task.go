package goodnewseveryone

import (
	"path/filepath"
	"os"
	"strings"
	"fmt"
	"io/ioutil"
	"errors"
	"encoding/json"
	"time"
)

var (
	errDuplicateTask = errors.New("Duplicate Task")
	errUnknownTask = errors.New("Unknown Task")
)

type TaskId string

type Tasks map[TaskId]Task

func configToTasks(log Log, configLoc string) (Tasks, error) {
	tasks := make(Tasks)
	err := filepath.Walk(configLoc, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "task.json") {
			log.Write(fmt.Sprintf("Task Config: %v", path))
			task, err := configToTask(configLoc, path)
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

func (tasks Tasks) Remove(taskId TaskId) error {
	if _, ok := tasks[taskId]; !ok {
		return errUnknownTask
	}
	if err := tasks[taskId].delete(); err != nil {
		return err
	}
	delete(tasks, taskId)
	return nil
}

func (tasks Tasks) Add(task Task) error {
	if _, ok := tasks[task.Id()]; ok {
		return errDuplicateTask
	}
	err := task.save()
	if err != nil {
		return err
	}
	tasks[task.Id()] = task
	return nil
}

type TaskType string

var (
	Sync = TaskType("sync")
	Backup = TaskType("backup")
	Move = TaskType("move")
)

var (
	errUndefinedTaskType = errors.New("Undefined Task Type: currently only sync, backup and move are supported")
)

type Task struct {
	Type TaskType
	Src LocationId
	Dst LocationId
	LastCompleted time.Time
}

func configToTask(configLoc string, filename string) (Task, error) {
	task := Task{}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return task, err
	}
	if err := json.Unmarshal(data, &task); err != nil {
		return task, err
	}
	if task.Type != Sync && task.Type != Backup && task.Type != Move {
		return task, errUndefinedTaskType
	}
	suffix := string(task.Id())+".complete"
	err = filepath.Walk(configLoc, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, suffix) {
			timeStr := strings.Replace(path, suffix, "", -1)
			t, err := time.Parse(DefaultTimeFormat, timeStr)
			if err != nil {
				return nil
			}
			if task.LastCompleted.Before(t) {
				task.LastCompleted = t
			}
		}
		return nil
	})
	if err != nil {
		return task, err
	}
	return task, nil
}

func NewTask(src LocationId, typ TaskType, dst LocationId) Task {
	return Task{
		Type: typ, 
		Src: src, 
		Dst: dst,
	}
}

func (this Task) filename() string {
	return fmt.Sprintf("%v.task.json", this.Id())
}

func (this Task) save() error {
	data, err := json.Marshal(this)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(this.filename(), data, 0666); err != nil {
		return err
	}
	return nil
}

func (this Task) delete() error {
	return os.Remove(this.filename())
}

func (this Task) String() string {
	return fmt.Sprintf("%v --%v-> %v", this.Src, this.Type, this.Dst)
}

func (this Task) Id() TaskId {
	return TaskId(fmt.Sprintf("%v-%v-%v", this.Src, this.Type, this.Dst))
}

func (this Task) Complete() {
	now := time.Now()
	err := ioutil.WriteFile(fmt.Sprintf("%v%v.complete", now.Format(DefaultTimeFormat), this.Id()), []byte{}, 0666)
	if err == nil {
		this.LastCompleted = now
	}
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
	case Move:
		return newMoveCommand(src.getLocal(), dst.getLocal()), nil
	}
	panic("unreachable")
}


