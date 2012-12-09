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

package web

import (
	"testing"
	"net/http"
	"goodnewseveryone/files"
	"goodnewseveryone"
	"bytes"
)

type writer struct {

}

func (this *writer) Header() http.Header {
	return nil
}

func (this *writer) Write([]byte) (int, error) {
	return 0, nil
}

func (this *writer) WriteHeader(int) {

}

func TestStatus(t *testing.T) {
	newWeb(goodnewseveryone.NewGNE(files.NewFiles(".")))
	r := bytes.NewBuffer(nil)
	req, err := http.NewRequest(".", "/", r)
	if err != nil {
		panic(err)
	}
	w := &writer{}
	this.handleStatus(w, req)
}
