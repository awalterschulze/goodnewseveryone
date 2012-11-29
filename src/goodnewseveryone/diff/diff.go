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
	"sort"
	"time"
	"goodnewseveryone/store"
	"path/filepath"
	"os"
)

type filemap map[string]bool

func (this filemap) list() []string {
	l := make([]string, 0, len(this))
	for filename, _ := range this {
		l = append(l, filename)
	}
	sort.Strings(l)
	return l
}

type fileList struct {
	Location string
	At time.Time
}

func newFileList(location string, at time.Time) *fileList {
	return &fileList{location, at}
}

type fileLists []*fileList

func (this fileLists) Len() int {
	return len(this)
}

func (this fileLists) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this fileLists) Less(i, j int) bool {
	return this[i].At.After(this[j].At)
}

func newFileLists(store store.FilelistStore) (fileLists, error) {
	locations, times, err := store.ListFilelists()
	if err != nil {
		return nil, err
	}
	lists := make(fileLists, len(locations))
	for i := range locations {
		lists[i] = newFileList(locations[i], times[i])
	}
	return lists, nil
}

func CreateFilelist(location string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func SaveFilelist(store store.FilelistStore, location string, at time.Time, files []string) error {
	return store.AddFilelist(location, at, files)
}

type Diff struct {
	store store.FilelistStore
	Previous time.Time
	Current time.Time
	Location string
}

func diffFilelist(oldList []string, newList []string) (created filemap, deleted filemap) {
	created = make(filemap)
	deleted = make(filemap)
	for _, newFile := range newList {
		created[newFile] = true
	}
	for _, oldFile := range oldList {
		if _, ok := created[oldFile]; ok {
			delete(created, oldFile)
		}
	}
	for _, oldFile := range oldList {
		deleted[oldFile] = true
	}
	for _, newFile := range newList {
		if _, ok := deleted[newFile]; ok {
			delete(deleted, newFile)
		}
	}
	return
}

func (this *Diff) Take() (created []string, deleted []string, err error) {
	prevList, err := this.store.ReadFilelist(this.Location, this.Previous)
	if err != nil {
		return nil, nil, err
	}
	curList, err := this.store.ReadFilelist(this.Location, this.Current)
	if err != nil {
		return nil, nil, err
	}
	createdList, deletedList := diffFilelist(prevList, curList)
	return createdList.list(), deletedList.list(), nil
}

type Diffs []*Diff

func (this Diffs) Len() int {
	return len(this)
}

func (this Diffs) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this Diffs) Less(i, j int) bool {
	return this[i].Current.After(this[j].Current)
}

type DiffsPerLocation map[string]Diffs

func NewDiffsPerLocation(store store.FilelistStore) (DiffsPerLocation, error) {
	filelists, err := newFileLists(store)
	if err != nil {
		return nil, err
	}
	diffs := make(DiffsPerLocation)
	sort.Sort(filelists)
	for _, filelist := range filelists {
		lastIndex := len(diffs[filelist.Location])-1
		if lastIndex == -1 {
			diffs[filelist.Location] = make(Diffs, 0)
		} else {
			diffs[filelist.Location][lastIndex].Previous = filelist.At
		}
		diffs[filelist.Location] = append(diffs[filelist.Location], &Diff{
			store: store,
			Current: filelist.At,
			Location: filelist.Location,
		})
	}
	for loc, _ := range diffs {
		if len(diffs[loc]) > 0 {
			diffs[loc] = diffs[loc][:len(diffs[loc])-1]
		}
		if len(diffs[loc]) == 0 {
			delete(diffs, loc)
		}
	}
	return diffs, nil
}
