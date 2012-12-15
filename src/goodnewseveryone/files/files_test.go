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
)

type data interface {
	Equal(that data) bool
}

type listFunc func(store store.Store) ([]string, error)
type readFunc func(store store.Store, name string) (data data, err error)
type addFunc func(store store.Store, name string, data data) error
type removeFunc func(store store.Store, name string) error

type testStore struct {
	list   listFunc
	read   readFunc
	add    addFunc
	remove removeFunc
}

func (this testStore) test(t *testing.T, name string, data data) {
	f := NewFiles(".")
	if err := this.add(f, name, data); err != nil {
		panic(err)
	}
	names, err := this.list(f)
	if err != nil {
		panic(err)
	}
	if len(names) != 1 {
		t.Fatalf("wrong number returned from list, expected 1, but got %v", len(names))
	}
	if names[0] != name {
		t.Fatalf("not the correct name, expected %v, but got %v", name, names[0])
	}
	thatData, err := this.read(f, name)
	if err != nil {
		panic(err)
	}
	if !data.Equal(thatData) {
		t.Fatalf("wrong data expected %#v, but got %#v", data, thatData)
	}
	if err := this.remove(f, name); err != nil {
		panic(err)
	}
	names, err = this.list(f)
	if err != nil {
		panic(err)
	}
	if len(names) != 0 {
		t.Fatalf("wrong number returned from list, expected 0, but got %v", len(names))
	}
}
