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
	"time"
	"os"
	"sync"
	"goodnewseveryone/store"
	"errors"
	"path/filepath"
	"strings"
	"io/ioutil"
	"encoding/json"
)

const defaultTimeFormat = time.RFC3339

var (
	ErrUnableToParseFilename = errors.New("unable to parse filename")
)

type filenameToNameFunc func(name string) (filename string)

func (this *files) list(suffix string, filenameToName filenameToNameFunc) (names []string, err error) {
	err = filepath.Walk(this.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, suffix) {
			names = append(names, filenameToName(path))
		}
		return nil
	})
	return names, err
}

func (this *files) read(filename string, unmarshal interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, unmarshal)
}

func (this *files) add(filename string, marshal interface{}) error {
	data, err := json.Marshal(marshal)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0666)
}

func (this *files) remove(filename string) error {
	return os.Remove(filename)
}

type files struct {
	sync.Mutex
	root string
	openLogFiles map[string]*os.File
	logFiles []string
}

func NewFiles(root string) store.Store {
	return &files{
		root: root,
		openLogFiles: make(map[string]*os.File),
		logFiles: findLogFiles(root),
	}
}
