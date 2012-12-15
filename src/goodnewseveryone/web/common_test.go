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
	"bytes"
	"fmt"
	"goodnewseveryone"
	"goodnewseveryone/files"
	"net/http"
	"net/url"
)

type writer struct {
}

func (this *writer) Header() http.Header {
	return make(http.Header)
}

func (this *writer) Write(data []byte) (int, error) {
	fmt.Printf("%v", string(data))
	return 0, nil
}

func (this *writer) WriteHeader(i int) {
	if i == http.StatusOK {
		return
	}
	panic(i)
}

func newRequest() *http.Request {
	req, err := http.NewRequest(".", "/", bytes.NewBuffer(nil))
	if err != nil {
		panic(err)
	}
	req.Form = make(url.Values)
	return req
}

func newHandles() (w http.ResponseWriter, r *http.Request) {
	g := goodnewseveryone.NewGNE(files.NewFiles("."))
	this = newWeb(g)
	go g.Start()
	return &writer{}, newRequest()
}
