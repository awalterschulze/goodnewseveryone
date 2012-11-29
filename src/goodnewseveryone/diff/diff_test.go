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

package diff

import (
	"testing"
	"goodnewseveryone/files"
	"time"
)

func TestDiff(t *testing.T) {
	f := files.NewFiles(".")
	diffs, err := NewDiffsPerLocation(f)
	if err != nil {
		panic(err)
	}
	location := "."
	locationName := "a"
	if len(diffs) != 0 {
		t.Fatalf("expected no diffs, but there is %v", diffs)
	}
	

	now1 := time.Now()
	files1, err := CreateFilelist(location)
	if err != nil {
		panic(err)
	}
	if err := SaveFilelist(f, locationName, now1, files1); err != nil {
		panic(err)
	}

	diffs, err = NewDiffsPerLocation(f)
	if err != nil {
		panic(err)
	}
	if len(diffs) != 0 {
		t.Fatalf("expected no diffs, but there is %v", diffs[location])
	}

	files2 := files1[1:]
	now2 := now1.Add(time.Hour)
	if err := SaveFilelist(f, locationName, now2, files2); err != nil {
		panic(err)
	}

	diffs, err = NewDiffsPerLocation(f)
	if err != nil {
		panic(err)
	}
	if len(diffs) != 1 {
		t.Fatalf("expected one diff, but there is %v", diffs[location])
	}

	if diff, ok := diffs[locationName]; !ok {
		t.Fatalf("no diff for location %v", locationName)
	} else {
		if len(diff) != 1 {
			t.Fatalf("expected one diff, but there is %v", diff)
		}
		created, deleted, err := diff[0].Take()
		if err != nil {
			panic(err)
		}
		if len(created) != 0 && len(deleted) != 1 {
			t.Fatalf("wrong diff")
		}
		if deleted[0] != files1[0] {
			t.Fatalf("expected deleted %v, but have %v", files1[0], deleted[0])
		}
	}

	files3 := files1
	now3 := now2.Add(time.Hour)
	if err := SaveFilelist(f, locationName, now3, files3); err != nil {
		panic(err)
	}

	diffs, err = NewDiffsPerLocation(f)
	if err != nil {
		panic(err)
	}
	if len(diffs) != 1 {
		t.Fatalf("expected one diff, but there is %v", diffs[location])
	}

	if diff, ok := diffs[locationName]; !ok {
		t.Fatalf("no diff for location %v", locationName)
	} else {
		if len(diff) != 2 {
			t.Fatalf("expected two diffs, but there is %v", diff)
		}
		created, deleted, err := diff[0].Take()
		if err != nil {
			panic(err)
		}
		if len(created) != 1 && len(deleted) != 0 {
			t.Fatalf("wrong diff %v created %v deleted", created, deleted)
		}
		if created[0] != files1[0] {
			t.Fatalf("expected created %v, but have %v", files1[0], created[0])
		}
	}

	ls, ts, err := f.ListFilelists()
	if err != nil {
		panic(err)
	}
	for i := range ls {
		if err := f.RemoveFilelist(ls[i], ts[i]); err != nil {
			panic(err)
		}
	}
}
