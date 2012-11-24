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
	"path/filepath"
	"os"
	"sort"
	"io/ioutil"
	"io"
	"strings"
	"time"
	"fmt"
	"errors"
)

type filelist map[string]bool

func (this filelist) list() []string {
	l := make([]string, 0, len(this))
	for filename, _ := range this {
		l = append(l, filename)
	}
	sort.Strings(l)
	return l
}

func (this filelist) writeTo(writer io.Writer) {
	for _, filename := range this.list() {
		writer.Write([]byte(filename+"\n"))
	}
}

func createList(locationKey string) (*os.File, error) {
	return os.Create(fmt.Sprintf("gne-_-%v-_-%v.list", locationKey, time.Now().Format(DefaultTimeFormat)))
}

func writeList(location Location) (error) {
	list, err := newFilelist(location.getLocal())
	if err != nil {
		return err
	}
	file, err := createList(string(location.Id()))
	if err != nil {
		return err
	}
	list.writeTo(file)
	return file.Close()
}

func newFilelist(location string) (filelist, error) {
	files := make(filelist)
	err := filepath.Walk(location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files[path] = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

type FileListFile struct {
	Filename string
	Location string
	At time.Time
}

func newFileListFile(filename string) (*FileListFile, error) {
	dataStr := strings.Replace(strings.Replace(filename, "gne-_-", "", 1), ".list", "", 1)
	dataStrs := strings.Split(dataStr, "-_-")
	if len(dataStrs) != 2 {
		return nil, errors.New("filename is not formatted correctly")
	}
	timeStr := dataStrs[1]
	location := dataStrs[0]
	t, err := time.Parse(DefaultTimeFormat, timeStr)
	if err != nil {
		return nil, err
	}
	return &FileListFile{filename, location, t}, nil
}

func readFilelist(reader io.Reader) (filelist, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	list := make(filelist)
	for _, line := range lines {
		list[line] = true
	}
	return list, nil
}

func (this *FileListFile) Read() ([]string, error) {
	file, err := os.Open(this.Filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	filelist, err := readFilelist(file)
	if err != nil {
		return nil, err
	}
	return filelist.list(), nil
}

type FileLists []*FileListFile

func (this FileLists) Len() int {
	return len(this)
}

func (this FileLists) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this FileLists) Less(i, j int) bool {
	return this[i].At.After(this[j].At)
}

func NewFileLists(root string) (FileLists, error) {
	filenames := []string{}
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".list") {
			filenames = append(filenames, path)
		}
		return nil
	})
	res := make(FileLists, len(filenames))
  	for i, filename := range filenames {
  		l, err := newFileListFile(filename)
  		if err != nil {
  			return nil, err
  		}
  		res[i] = l
  	}
  	return res, nil
}

type Diff struct {
	At time.Time
	CurrentFilename string
	PreviousFilename string
}

func diffFilelist(oldList filelist, newList filelist) (created filelist, deleted filelist) {
	created = make(filelist)
	deleted = make(filelist)
	for newFile, _ := range newList {
		if _, ok := oldList[newFile]; !ok {
			created[newFile] = true
		}
	}
	for oldFile, _ := range oldList {
		if _, ok := newList[oldFile]; !ok {
			deleted[oldFile] = true
		}
	}
	return
}

func (this *Diff) Take() (created []string, deleted []string, err error) {
	curfile, err := os.Open(this.CurrentFilename)
	if err != nil {
		return nil, nil, err
	}
	defer curfile.Close()
	curList, err := readFilelist(curfile)
	if err != nil {
		return nil, nil, err
	}
	prevfile, err := os.Open(this.PreviousFilename)
	if err != nil {
		return nil, nil, err
	}
	defer prevfile.Close()
	prevList, err := readFilelist(prevfile)
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
	return this[i].At.After(this[j].At)
}

type DiffsPerLocation map[string]Diffs

func NewDiffsPerLocation(root string) (DiffsPerLocation, error) {
	filelists, err := NewFileLists(root)
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
			diffs[filelist.Location][lastIndex].PreviousFilename = filelist.Filename
		}
		diffs[filelist.Location] = append(diffs[filelist.Location], &Diff{
			CurrentFilename: filelist.Filename,
			At: filelist.At,
		})
	}
	for loc, _ := range diffs {
		if len(diffs[loc]) > 0 {
			diffs[loc] = diffs[loc][:len(diffs[loc])-1]
		}
	}
	return diffs, nil
}
