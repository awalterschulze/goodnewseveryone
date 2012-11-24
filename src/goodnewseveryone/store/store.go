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

package store

import (
	"errors"
	"time"
)

type Store interface{
	LogStore
	TaskStore
	LocationStore
	ConfigStore
	FilelistStore
}

var (
	ErrLogSessionAlreadyExists = errors.New("Log Session with key already exists")
	ErrLogSessionDoesNotExist = errors.New("Log Session with key does not exist")
	ErrLogSessionIsOpenCannotDelete = errors.New("Log Session is open, it cannot be deleted")
)

type LogStore interface {
	NewLogSession(key time.Time) error
	ListLogSessions() []time.Time
	ReadFromLogSession(key time.Time) ([]time.Time, []string, error)
	WriteToLogSession(key time.Time, line string) error
	DeleteLogSession(key time.Time) error
	CloseLogSession(key time.Time) error
}

var (
	ErrTaskTypeDoesNotExist = errors.New("Task Type does not exist")
	ErrTaskAlreadyExists = errors.New("Task already exists")
	ErrTaskDoesNotExist = errors.New("Task does not exist")
)

type TaskStore interface {
	ListTaskTypes() (names []string, err error)
	ReadTaskType(name string) (cmdStr string, err error)
	AddTaskType(name string, cmdStr string) error
	RemoveTaskType(name string) error

	ListTasks() (taskIds []string, err error)
	ReadTask(taskId string) (src, taskType, dst string, err error)
	AddTask(taskId string, src, taskType, dst string) error
	RemoveTask(taskId string) error

	ListTaskCompleted(taskId string) ([]time.Time, error)
	AddTaskCompleted(taskId string, now time.Time) error
	RemoveTaskCompleted(taskId string, then time.Time) error
}

var (
	ErrRemoteLocationTypeDoesNotExist = errors.New("Remote Location Type does not exist")
)

type LocationStore interface {
	ListLocalLocations() (names []string, err error)
	ReadLocalLocation(name string) (local string, err error)
	AddLocalLocation(name string, local string) error
	RemoveLocalLocation(name string) error

	ListRemoteLocationTypes() (names []string, err error)
	ReadRemoteLocationType(name string) (mount string, unmount string, err error)
	AddRemoteLocationType(name string, mount string, unmount string) error
	RemoveRemoteLocationType(name string) error

	ListRemoteLocations() (names []string, err error)
	ReadRemoteLocation(name string) (typ string, ipAddress string, username string, password string, remote string, err error)
	AddRemoteLocation(name string, typ string, ipAddress string, username string, password string, remote string) error
	RemoveRemoteLocation(name string) error
}

type FilelistStore interface {
	ListFilelists() (locations []string, times []time.Time, err error)
	ReadFilelist(location string, t time.Time) ([]string, error)
	AddFilelist(location string, t time.Time, files []string) error
	RemoveFilelist(location string, t time.Time) error
}

type ConfigStore interface {
	ResetWaitTime() error
	GetWaitTime() (time.Duration, error)
	SetWaitTime(w time.Duration) error
	ResetMountFolder() error
	GetMountFolder() (string, error)
	SetMountFolder(folder string) error
}

