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
	"goodnewseveryone/store"
	"testing"
	"reflect"
)

func (this *local) Equal(that data) bool {
	return that.(*local).Local == this.Local
}

func TestLocalLocation(t *testing.T) {
	testStore{
		list: func(store store.Store) ([]string, error) {
			return store.ListLocalLocations()
		},
		read: func(store store.Store, name string) (data data, err error) {
			l, err := store.ReadLocalLocation(name)
			return &local{l}, err
		},
		add: func(store store.Store, name string, data data) error {
			return store.AddLocalLocation(name, data.(*local).Local)
		},
		remove: func(store store.Store, name string) error {
			return store.RemoveLocalLocation(name)
		},
	}.test(t, "abc", &local{"hello"})
}

func (this *remoteType) Equal(that data) bool {
	r := that.(*remoteType)
	return reflect.DeepEqual(this, r)
}

func TestRemoteLocationType(t *testing.T) {
	testStore{
		list: func(store store.Store) ([]string, error) {
			return store.ListRemoteLocationTypes()
		},
		read: func(store store.Store, name string) (data data, err error) {
			m, u, err := store.ReadRemoteLocationType(name)
			return &remoteType{m, u}, err
		},
		add: func(store store.Store, name string, data data) error {
			return store.AddRemoteLocationType(name, data.(*remoteType).Mount, data.(*remoteType).Unmount)
		},
		remove: func(store store.Store, name string) error {
			return store.RemoveRemoteLocationType(name)
		},
	}.test(t, "abc", &remoteType{"mount", "unmount"})
}

func (this *remote) Equal(that data) bool {
	r := that.(*remote)
	return reflect.DeepEqual(this, r)
}

func TestRemoteLocation(t *testing.T) {
	testStore{
		list: func(store store.Store) ([]string, error) {
			return store.ListRemoteLocations()
		},
		read: func(store store.Store, name string) (data data, err error) {
			typ, ipAddress, username, password, remoteFolder, err := store.ReadRemoteLocation(name)
			r := &remote{
				Type: typ,
				IPAddress: ipAddress,
				Username: username,
				Password: password,
				Remote: remoteFolder,
			}
			return r, err
		},
		add: func(store store.Store, name string, data data) error {
			r := data.(*remote)
			return store.AddRemoteLocation(name, r.Type, r.IPAddress, r.Username, r.Password, r.Remote)
		},
		remove: func(store store.Store, name string) error {
			return store.RemoveRemoteLocation(name)
		},
	}.test(t, "SharedFolder", &remote{
		Type: "smb",
		IPAddress: "192.168.0.1",
		Username: "walter",
		Password: "schulze",
		Remote: "Share",
	})
}