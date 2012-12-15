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

package location

import (
	"fmt"
	"goodnewseveryone/command"
	"goodnewseveryone/log"
	gstore "goodnewseveryone/store"
	"strings"
)

type Locations map[string]Location

type Store interface {
	gstore.LocationStore
	gstore.ConfigStore
}

type RemoteLocationType struct {
	Name    string
	Mount   string
	Unmount string
}

func ListRemoteLocationTypes(store Store) ([]RemoteLocationType, error) {
	remoteTypes, err := store.ListRemoteLocationTypes()
	if err != nil {
		return nil, err
	}
	remotes := make([]RemoteLocationType, 0, len(remoteTypes))
	for _, t := range remoteTypes {
		mount, unmount, err := store.ReadRemoteLocationType(t)
		if err != nil {
			return nil, err
		}
		remotes = append(remotes, RemoteLocationType{
			Name:    t,
			Mount:   mount,
			Unmount: unmount,
		})
	}
	return remotes, nil
}

func NewLocations(log log.Log, store Store) (Locations, error) {
	locations := make(Locations)
	locals, err := store.ListLocalLocations()
	if err != nil {
		return nil, err
	}
	for _, name := range locals {
		l, err := store.ReadLocalLocation(name)
		if err != nil {
			log.Error(err)
			continue
		}
		loc := NewLocalLocation(name, l)
		if err := locations.Add(store, loc); err != nil {
			log.Error(err)
			continue
		}
		log.Write(fmt.Sprintf("Location Configured: %v", loc))
	}
	mountFolder, err := store.GetMountFolder()
	if err != nil {
		return nil, err
	}
	remoteTypes, err := ListRemoteLocationTypes(store)
	if err != nil {
		return nil, err
	}
	remotes, err := store.ListRemoteLocations()
	if err != nil {
		return nil, err
	}
	for _, name := range remotes {
		typ, ipAddress, username, password, remoteFolder, err := store.ReadRemoteLocation(name)
		if err != nil {
			log.Error(err)
			continue
		}
		mount, unmount := "", ""
		for i, rtype := range remoteTypes {
			if rtype.Name == typ {
				mount, unmount = remoteTypes[i].Mount, remoteTypes[i].Unmount
				break
			}
		}
		if len(mount) == 0 || len(unmount) == 0 {
			log.Error(gstore.ErrRemoteLocationTypeDoesNotExist)
			continue
		}
		loc := NewRemoteLocation(name, typ, ipAddress, username, password, remoteFolder, mountFolder, mount, unmount)
		if err := locations.Add(store, loc); err != nil {
			log.Error(err)
			continue
		}
		log.Write(fmt.Sprintf("Location Configured: %v", loc))
	}
	return locations, nil
}

func (locations Locations) Remove(store Store, locId string) error {
	if _, ok := locations[locId]; !ok {
		return gstore.ErrLocationDoesNotExist
	}
	if err := locations[locId].delete(store); err != nil {
		return err
	}
	delete(locations, locId)
	return nil
}

func (locations Locations) Add(store Store, loc Location) error {
	if _, ok := locations[loc.Id()]; ok {
		return gstore.ErrLocationAlreadyExists
	}
	err := loc.save(store)
	if err != nil {
		return err
	}
	locations[loc.Id()] = loc
	return nil
}

func (locations Locations) String() string {
	locs := make([]string, 0, len(locations))
	for _, loc := range locations {
		locs = append(locs, loc.String())
	}
	return "[" + strings.Join(locs, ", ") + "]"
}

type Location interface {
	String() string
	Id() string
	NewLocatedCommand() command.Command
	Located(log log.Log, output string) bool
	NewPreparedCommand() command.Command
	Prepared(log log.Log, output string) bool
	NewPrepareCommand() command.Command
	NewMountCommand() command.Command
	NewUmountCommand() command.Command
	GetLocal() string
	save(store Store) error
	delete(store Store) error
}

type RemoteLocation struct {
	Name        string
	Type        string
	IPAddress   string
	Username    string
	Password    string
	Remote      string
	MountFolder string
	Mount       string
	Unmount     string
}

func NewRemoteLocation(name string, typ string, ipaddress string, username string, password string, remote string, mountFolder string, mount, unmount string) *RemoteLocation {
	return &RemoteLocation{
		name,
		typ,
		ipaddress,
		username,
		password,
		remote,
		mountFolder,
		mount,
		unmount,
	}
}

func (this *RemoteLocation) String() string {
	return "REMOTE=" + this.IPAddress + "_" + string(this.Type) + "//" + this.Remote
}

func (this *RemoteLocation) Id() string {
	return this.Name
}

func (this *RemoteLocation) NewLocatedCommand() command.Command {
	return command.NewNMap(this.IPAddress)
}

func (this *RemoteLocation) Located(log log.Log, output string) bool {
	if !strings.Contains(output, "Host is up") {
		log.Write(fmt.Sprintf("Cannot Locate %v", this))
		return false
	}
	return true
}

func (this *RemoteLocation) NewPreparedCommand() command.Command {
	return command.NewLS(this.GetLocal())
}

func (this *RemoteLocation) Prepared(log log.Log, output string) bool {
	if strings.Contains(output, "No such file or directory") {
		return false
	}
	return true
}

func (this *RemoteLocation) NewPrepareCommand() command.Command {
	return command.NewMkdir(this.GetLocal())
}

func (this *RemoteLocation) NewMountCommand() command.Command {
	return command.NewMount(this.Mount, this.IPAddress, this.Username, this.Password, this.Remote, this.GetLocal())
}

func (this *RemoteLocation) NewUmountCommand() command.Command {
	return command.NewUnmount(this.Unmount, this.GetLocal())
}

func (this *RemoteLocation) GetLocal() string {
	return this.MountFolder + "/" + string(this.Id())
}

func (this *RemoteLocation) save(store Store) error {
	return store.AddRemoteLocation(this.Name, this.Type, this.IPAddress, this.Username, this.Password, this.Remote)
}

func (this *RemoteLocation) delete(store Store) error {
	return store.RemoveRemoteLocation(this.Name)
}

type LocalLocation struct {
	Name  string
	Local string
}

func NewLocalLocation(name, local string) *LocalLocation {
	return &LocalLocation{name, local}
}

func (this *LocalLocation) String() string {
	return "LOCAL=" + this.Local
}

func (this *LocalLocation) Id() string {
	return this.Name
}

func (this *LocalLocation) NewLocatedCommand() command.Command {
	return nil
}

func (this *LocalLocation) Located(log log.Log, output string) bool {
	return true
}

func (this *LocalLocation) NewPreparedCommand() command.Command {
	return nil
}

func (this *LocalLocation) Prepared(log log.Log, output string) bool {
	return true
}

func (this *LocalLocation) NewPrepareCommand() command.Command {
	return nil
}

func (this *LocalLocation) NewMountCommand() command.Command {
	return nil
}

func (this *LocalLocation) NewUmountCommand() command.Command {
	return nil
}

func (this *LocalLocation) GetLocal() string {
	return this.Local
}

func (this *LocalLocation) save(store Store) error {
	return store.AddLocalLocation(this.Name, this.Local)
}

func (this *LocalLocation) delete(store Store) error {
	return store.RemoveLocalLocation(this.Name)
}
