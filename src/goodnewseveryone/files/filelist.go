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
	"io/ioutil"
	"fmt"
	"strings"
	"os"
	"errors"
)

var filelistSuffix = ".filelist.txt"

func filelistNameToFilename(name string) (filename string) {
	return fmt.Sprintf("%v%v", name, filelistSuffix)
}

func filenameTofilelistName(filename string) (name string) {
	return strings.Replace(filename, filelistSuffix, "", 1)
}

func filelistNameToLocationandTime(name string) (location string, t time.Time, err error) {
	ss := strings.Split(name, "---")
	if len(ss) != 2 {
		return "", t, errors.New("filelist filename parse error")
	}
	t, err = time.Parse(defaultTimeFormat, ss[1])
	if err != nil {
		return "", t, err
	}
	return ss[0], t, nil
}

func locationAndTimeToFilelistName(location string, t time.Time) string {
	return fmt.Sprintf("%v---%v", location, t.Format(defaultTimeFormat))
}

func (this *files) ListFilelists() (locations []string, times []time.Time, err error) {
	this.Lock()
	defer this.Unlock()
	names, err := this.list(filelistSuffix, filenameTofilelistName)
	if err != nil {
		return nil, nil, err
	}
	locations = make([]string, 0, len(names))
	times = make([]time.Time, 0, len(names))
	for _, name := range names {
		l, t, err := filelistNameToLocationandTime(name)
		if err != nil {
			continue
		}
		locations = append(locations, l)
		times = append(times, t)
	}
	return locations, times, nil
}
	
func (this *files) ReadFilelist(location string, t time.Time) ([]string, error) {
	this.Lock()
	defer this.Unlock()
	filename := filelistNameToFilename(locationAndTimeToFilelistName(location, t))
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	return lines, nil
}
	
func (this *files) AddFilelist(location string, t time.Time, files []string) error {
	this.Lock()
	defer this.Unlock()
	filename := filelistNameToFilename(locationAndTimeToFilelistName(location, t))
	data := []byte(strings.Join(files, "\n"))
	return ioutil.WriteFile(filename, data, 0666)
}
	
func (this *files) RemoveFilelist(location string, t time.Time) error {
	this.Lock()
	defer this.Unlock()
	filename := filelistNameToFilename(locationAndTimeToFilelistName(location, t))
	return os.Remove(filename)
}
