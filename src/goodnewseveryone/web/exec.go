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
	"net/http"
)

func init() {
	http.HandleFunc("/unblock", func(w http.ResponseWriter, r *http.Request) {
		this.handleUnblock(w,r)
	})
	http.HandleFunc("/stopandblock", func(w http.ResponseWriter, r *http.Request) {
		this.handleStopAndBlock(w,r)
	})
	http.HandleFunc("/now", func(w http.ResponseWriter, r *http.Request) {
		this.handleNow(w,r)
	})
}

func (this *web) handleUnblock(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	this.gne.Unblock()
	redirectHomeTemplate.Execute(w, quickHome)
	footerTemplate.Execute(w, nil)
}

func (this *web) handleStopAndBlock(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	this.gne.StopAndBlock()
	redirectHomeTemplate.Execute(w, quickHome)
	footerTemplate.Execute(w, nil)
}

func (this *web) handleNow(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	taskName := r.FormValue("name")
	this.gne.Now(taskName)
	redirectHomeTemplate.Execute(w, quickHome)
	footerTemplate.Execute(w, nil)
}
