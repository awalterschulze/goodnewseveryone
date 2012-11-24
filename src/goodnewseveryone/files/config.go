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
//limitations under the License.package files

package files

import (
	"time"
)

type waittime struct {
	Duration time.Duration
}

func waittimeFilename(s string) string {
	return "waittime.config.json"
}

func (this *files) ResetWaitTime() error {
	this.Lock()
	defer this.Unlock()
	names, err := this.list(waittimeFilename(""), waittimeFilename)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}
	return this.remove(waittimeFilename(""))
}

func (this *files) GetWaitTime() (time.Duration, error) {
	this.Lock()
	defer this.Unlock()
	names, err := this.list(waittimeFilename(""), waittimeFilename)
	if err != nil {
		return time.Duration(0), err
	}
	if len(names) == 0 {
		return time.Duration(0), nil
	}
	w := &waittime{}
	if err := this.read(waittimeFilename(""), w); err != nil {
		return w.Duration, err
	}
	return w.Duration, nil
}

func (this *files) SetWaitTime(w time.Duration) error {
	this.Lock()
	defer this.Unlock()
	return this.add(waittimeFilename(""), &waittime{w})
}

type mountFolder struct {
	MountFolder string
}

func mountFolderFilename(s string) string {
	return "mountfolder.config.json"
}

func (this *files) ResetMountFolder() error {
	this.Lock()
	defer this.Unlock()
	names, err := this.list(mountFolderFilename(""), mountFolderFilename)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}
	return this.remove(mountFolderFilename(""))
}

func (this *files) GetMountFolder() (string, error) {
	this.Lock()
	defer this.Unlock()
	names, err := this.list(mountFolderFilename(""), mountFolderFilename)
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", nil
	}
	m := &mountFolder{}
	if err := this.read(mountFolderFilename(""), m); err != nil {
		return "", err
	}
	return m.MountFolder, nil
}

func (this *files) SetMountFolder(folder string) error {
	this.Lock()
	defer this.Unlock()
	return this.add(mountFolderFilename(""), &mountFolder{folder})
}
