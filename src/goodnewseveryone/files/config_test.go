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
	"testing"
	"time"
)

func TestWaitTime(t *testing.T) {
	f := NewFiles(".")
	if err := f.ResetWaitTime(); err != nil {
		panic(err)
	}
	w, err := f.GetWaitTime()
	if err != nil {
		panic(err)
	}
	if int(w) != 0 {
		t.Fatalf("Non default value WaitTime = %v", w)
	}
	d := time.Second
	if err := f.SetWaitTime(d); err != nil {
		panic(err)
	}
	w, err = f.GetWaitTime()
	if w != d {
		t.Fatalf("expected %v, but got %v", d, w)
	}
	if err := f.ResetWaitTime(); err != nil {
		panic(err)
	}
}

func TestMountFolder(t *testing.T) {
	f := NewFiles(".")
	if err := f.ResetMountFolder(); err != nil {
		panic(err)
	}
	def, err := f.GetMountFolder()
	if err != nil {
		panic(err)
	}
	folder := "/media"
	if err := f.SetMountFolder(folder); err != nil {
		panic(err)
	}
	folder2, err := f.GetMountFolder()
	if folder != folder2 {
		t.Fatalf("expected %v, but got %v", folder, folder2)
	}
	if err := f.ResetMountFolder(); err != nil {
		panic(err)
	}
	def2, err := f.GetMountFolder()
	if err != nil {
		panic(err)
	}
	if def != def2 {
		t.Fatalf("expected %v, but got %v", def, def2)
	}
}
