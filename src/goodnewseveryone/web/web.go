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
	gne "goodnewseveryone"
	"net/http"
)

var (
	this = &web{}
)

type web struct {
	gne gne.GNE
}

func newWeb(gne gne.GNE) *web {
	this.gne = gne
	return this
}

func Serve(gne gne.GNE, port string) {
	this = newWeb(gne)
	http.ListenAndServe(":"+port, nil)
}
