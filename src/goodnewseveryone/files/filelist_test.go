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
	"reflect"
)

func TestFilelist(t *testing.T) {
	f := NewFiles(".")
	ls, ts, err := f.ListFilelists()
	if err != nil {
		panic(err)
	}
	if len(ls) != 0 {
		t.Fatalf("some locations listed")
	}
	if len(ts) != 0 {
		t.Fatalf("some times listed")
	}

	location := "MySharedFolder"
	now := time.Now()
	filelist := []string{"/file1, /file2.txt"}

	if err := f.AddFilelist(location, now, filelist); err != nil {
		panic(err)
	}

	ls, ts, err = f.ListFilelists()
	if err != nil {
		panic(err)
	}
	if len(ls) != 1 {
		t.Fatalf("not 1 location listed, but %v", len(ls))
	}
	if len(ts) != 1 {
		t.Fatalf("not 1 time listed, but $v", len(ts))
	}

	filelist2, err := f.ReadFilelist(ls[0], ts[0])
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(filelist2, filelist) {
		t.Fatalf("filelist not equal %v, but expected %v", filelist2, filelist)
	}

	if err := f.RemoveFilelist(location, now); err != nil {
		panic(err)
	}

	ls, ts, err = f.ListFilelists()
	if err != nil {
		panic(err)
	}
	if len(ls) != 0 {
		t.Fatalf("some locations listed")
	}
	if len(ts) != 0 {
		t.Fatalf("some times listed")
	}
}