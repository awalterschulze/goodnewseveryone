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
	"path/filepath"
	"fmt"
	"strings"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

func taskTypeNameToFilename(taskTypeName string) (filename string) {
	return fmt.Sprintf("%v.tasktype.json", taskTypeName)
}

func filenameToTaskTypeName(filename string) (taskTypeName string) {
	return strings.Replace(filename, ".tasktype.json", "", 1)
}

type taskType struct {
	Name string
	TaskType string
}

func (this *files) ListTaskTypes() (names []string, err error) {
	this.Lock()
	defer this.Unlock()
	return this.list("tasktype.json", filenameToTaskTypeName)
}

func (this *files) ReadTaskType(name string) (string, error) {
	this.Lock()
	defer this.Unlock()
	filename := taskTypeNameToFilename(name)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	t := taskType{}
	if err := json.Unmarshal(data, &t); err != nil {
		return "", err
	}
	return t.TaskType, nil
}
	
func (this *files) AddTaskType(name string, tt string) error {
	this.Lock()
	defer this.Unlock()
	filename := taskTypeNameToFilename(name)
	t := taskType{
		Name: name,
		TaskType: tt,
	}
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, data, 0666); err != nil {
		return err
	}
	return nil
}
	
func (this *files) RemoveTaskType(name string) error {
	this.Lock()
	defer this.Unlock()
	return os.Remove(taskTypeNameToFilename(name))
}

func taskIdToFilename(taskId string) (filename string) {
	return fmt.Sprintf("%v.task.json", taskId)
}

func filenameToTaskId(filename string) (taskId string) {
	return strings.Replace(filename, ".task.json", "", 1)
}

type task struct {
	Src string
	Typ string
	Dst string
}

func (this *files) ListTasks() (taskIds []string, err error) {
	this.Lock()
	defer this.Unlock()
	return this.list("task.json", filenameToTaskId)
}

func (this *files) ReadTask(taskId string) (src, taskType, dst string, err error) {
	this.Lock()
	defer this.Unlock()
	data, err := ioutil.ReadFile(taskIdToFilename(taskId))
	if err != nil {
		return "", "", "", err
	}
	t := &task{}
	if err := json.Unmarshal(data, &t); err != nil {
		return "", "", "", err
	}
	return t.Src, t.Typ, t.Dst, nil
}
	
func (this *files) AddTask(taskId string, src, taskType, dst string) error {
	this.Lock()
	defer this.Unlock()
	t := task{
		Src: src,
		Typ: taskType,
		Dst: dst,
	}
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	filename := taskIdToFilename(taskId)
	if err := ioutil.WriteFile(filename, data, 0666); err != nil {
		return err
	}
	return nil
}
	
func (this *files) RemoveTask(taskId string) error {
	this.Lock()
	defer this.Unlock()
	return os.Remove(taskIdToFilename(taskId))
}

func taskAndTimeToFilename(taskId string, t time.Time) (filename string) {
	return fmt.Sprintf("%v---%v.complete", taskId, t.Format(defaultTimeFormat))
}

func filenameToTaskAndTime(filename string) (taskId string, t time.Time, err error) {
	filename = strings.Replace(filename, ".complete", "", 1)
	ss := strings.Split(filename, "---")
	if len(ss) != 2 {
		return "", time.Time{}, ErrUnableToParseFilename
	}
	taskId = ss[0]
	t, err = time.Parse(defaultTimeFormat, ss[1])
	return taskId, t, err
}

func (this *files) ListTaskCompleted(taskId string) (times []time.Time, err error) {
	this.Lock()
	defer this.Unlock()
	err = filepath.Walk(this.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".complete") {
			fileTaskId, t, err := filenameToTaskAndTime(path)
			if err != nil {
				return err
			}
			if taskId == fileTaskId {
				times = append(times, t)
			}
		}
		return nil
	})
	return times, err
}
	
func (this *files) AddTaskCompleted(taskId string, now time.Time) error {
	this.Lock()
	defer this.Unlock()
	filename := taskAndTimeToFilename(taskId, now)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	return f.Close()
}
	
func (this *files) RemoveTaskCompleted(taskId string, then time.Time) error {
	this.Lock()
	defer this.Unlock()
	filename := taskAndTimeToFilename(taskId, then)
	return os.Remove(filename)
}
