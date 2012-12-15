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
	"goodnewseveryone/files"
	"reflect"
	"testing"
)

type loggy struct {
}

func (this *loggy) Write(str string) {
	fmt.Printf(str + "\n")
}

func (this *loggy) Run(name string, arg ...string) {
	panic(name)
}

func (this *loggy) Error(err error) {
	panic(err)
}

func (this *loggy) Output(output []byte) {
	this.Write(string(output))
}

func (this *loggy) Close() {

}

func TestLocations(t *testing.T) {
	store := files.NewFiles(".")
	mount := "mount -o username=%v,password=%v,nounix,noserverino,sec=ntlmssp -t cifs //%v/%v %v"
	unmount := "umount -l %v"
	if err := store.AddRemoteLocationType("smb", mount, unmount); err != nil {
		panic(err)
	}
	if err := store.SetMountFolder("/media"); err != nil {
		panic(err)
	}
	thelog := &loggy{}
	locations, err := NewLocations(thelog, store)
	if err != nil {
		panic(err)
	}
	local := &LocalLocation{"LocalName", "LocalLocal"}
	locations.Add(store, local)
	remote := &RemoteLocation{
		Name:        "RemoteName",
		Type:        "smb",
		IPAddress:   "192.168.0.1",
		Username:    "walter",
		Password:    "schulze",
		Remote:      "MySharedFolder",
		MountFolder: "/media",
		Mount:       mount,
		Unmount:     unmount,
	}
	locations.Add(store, remote)
	localEqual, remoteEqual := false, false
	for _, location := range locations {
		if r, ok := location.(*RemoteLocation); ok {
			if !reflect.DeepEqual(r, remote) {
				t.Fatalf("%#v != %#v", r, remote)
			}
			remoteEqual = true
		} else if l, ok := location.(*LocalLocation); ok {
			if !reflect.DeepEqual(l, local) {
				t.Fatalf("%v != %v", l, local)
			}
			localEqual = true
		} else {
			t.Fatalf("unknown type %T", location)
		}
	}
	if !localEqual || !remoteEqual {
		t.Fatalf("local or remote is not equal")
	}
	locations, err = NewLocations(thelog, store)
	if err != nil {
		panic(err)
	}
	localEqual, remoteEqual = false, false
	for _, location := range locations {
		if r, ok := location.(*RemoteLocation); ok {
			if !reflect.DeepEqual(r, remote) {
				t.Fatalf("%v != %v", r, remote)
			}
			remoteEqual = true
		} else if l, ok := location.(*LocalLocation); ok {
			if !reflect.DeepEqual(l, local) {
				t.Fatalf("%v != %v", l, local)
			}
			localEqual = true
		} else {
			t.Fatalf("unknown type %T", location)
		}
	}
	if !localEqual || !remoteEqual {
		t.Fatalf("local or remote is not equal")
	}
	if err := locations.Remove(store, local.Id()); err != nil {
		panic(err)
	}
	if len(locations) != 1 {
		t.Fatalf("location not removed")
	}
	if err := locations.Remove(store, remote.Id()); err != nil {
		panic(err)
	}
	if len(locations) != 0 {
		t.Fatalf("locations not removed")
	}
	locations, err = NewLocations(thelog, store)
	if err != nil {
		panic(err)
	}
	if len(locations) != 0 {
		t.Fatalf("locations really not removed")
	}
	if err := store.ResetMountFolder(); err != nil {
		panic(err)
	}
	if err := store.RemoveRemoteLocationType("smb"); err != nil {
		panic(err)
	}
}
