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
	"fmt"
	"strings"
)

var localFileSuffix = ".local.json"

type local struct {
	Local string
}

func localLocationNameToFilename(name string) (filename string) {
	return fmt.Sprintf("%v%v", name, localFileSuffix)
}

func filenameToLocalLocationName(filename string) (name string) {
	return strings.Replace(filename, localFileSuffix, "", 1)
}

func (this *files) ListLocalLocations() (names []string, err error) {
	this.Lock()
	defer this.Unlock()
	return this.list(localFileSuffix, filenameToLocalLocationName)
}

func (this *files) ReadLocalLocation(name string) (localFolder string, err error) {
	this.Lock()
	defer this.Unlock()
	l := &local{}
	if err := this.read(localLocationNameToFilename(name), l); err != nil {
		return "", err
	}
	return l.Local, nil
}

func (this *files) AddLocalLocation(name string, localFolder string) error {
	this.Lock()
	defer this.Unlock()
	return this.add(localLocationNameToFilename(name), &local{localFolder})
}

func (this *files) RemoveLocalLocation(name string) error {
	this.Lock()
	defer this.Unlock()
	return this.remove(localLocationNameToFilename(name))
}

var remoteTypeFileSuffix = ".remotetype.json"

type remoteType struct {
	Mount   string
	Unmount string
}

func remoteLocationTypeToFilename(name string) (filename string) {
	return fmt.Sprintf("%v%v", name, remoteTypeFileSuffix)
}

func filenameToRemoteLocationType(filename string) (name string) {
	return strings.Replace(filename, remoteTypeFileSuffix, "", 1)
}

func (this *files) ListRemoteLocationTypes() (names []string, err error) {
	this.Lock()
	defer this.Unlock()
	return this.list(remoteTypeFileSuffix, filenameToRemoteLocationType)
}

func (this *files) ReadRemoteLocationType(name string) (mount string, unmount string, err error) {
	this.Lock()
	defer this.Unlock()
	r := &remoteType{}
	if err := this.read(remoteLocationTypeToFilename(name), r); err != nil {
		return "", "", err
	}
	return r.Mount, r.Unmount, nil
}

func (this *files) AddRemoteLocationType(name string, mount string, unmount string) error {
	this.Lock()
	defer this.Unlock()
	return this.add(remoteLocationTypeToFilename(name), &remoteType{mount, unmount})
}

func (this *files) RemoveRemoteLocationType(name string) error {
	this.Lock()
	defer this.Unlock()
	return this.remove(remoteLocationTypeToFilename(name))
}

var remoteFileSuffix = ".remote.json"

type remote struct {
	Type      string
	IPAddress string
	Username  string
	Password  string
	Remote    string
}

func remoteLocationNameToFilename(name string) (filename string) {
	return fmt.Sprintf("%v%v", name, remoteFileSuffix)
}

func filenameToRemoteLocationName(filename string) (name string) {
	return strings.Replace(filename, remoteFileSuffix, "", 1)
}

func (this *files) ListRemoteLocations() (names []string, err error) {
	this.Lock()
	defer this.Unlock()
	return this.list(remoteFileSuffix, filenameToRemoteLocationName)
}

func (this *files) ReadRemoteLocation(name string) (typ string, ipAddress string, username string, password string, remoteFolder string, err error) {
	this.Lock()
	defer this.Unlock()
	r := &remote{}
	if err := this.read(remoteLocationNameToFilename(name), r); err != nil {
		return "", "", "", "", "", err
	}
	return r.Type, r.IPAddress, r.Username, r.Password, r.Remote, nil
}

func (this *files) AddRemoteLocation(name string, typ string, ipAddress string, username string, password string, remoteFolder string) error {
	this.Lock()
	defer this.Unlock()
	return this.add(remoteLocationNameToFilename(name), &remote{typ, ipAddress, username, password, remoteFolder})
}

func (this *files) RemoveRemoteLocation(name string) error {
	this.Lock()
	defer this.Unlock()
	return this.remove(remoteLocationNameToFilename(name))
}
