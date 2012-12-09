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
	"text/template"
	"fmt"
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		this.handleStatus(w,r)
	})
}

var (
	statusTemplate = template.Must(template.New("status").Parse(`
		<div>{{if .IsRunning}}Running{{else}}Not Running{{if .IsReady}}<a href="./now">Now</a>{{end}}{{end}}</div>
		<div>{{if .IsReady}}Ready<a href="./stopandblock">StopAndBlock</a>{{else}}Blocked<a href="./unblock">Unblock</a>{{end}}</div>
		<div>WaitTime {{.GetWaitTime}}<a href="./waittime">Set</a></div>
		<div><a href="./man">Task Management</a></div>
		<div><a href="./diffs">Diffs</a></div>
		<div><a href=".">Current Status</a></div>
	`))
)

func (this *web) handleStatus(w http.ResponseWriter, r *http.Request) {
	min := r.FormValue("min")
	max := r.FormValue("max")
	headerTemplate.Execute(w, nil)
	redirectHomeTemplate.Execute(w, &redirectHome{
		Min: min,
		Max: max,
		Delay: slow,
	})
	statusTemplate.Execute(w, this.gne)
	logs, err := this.newLogs(min, max)
	if err != nil {
		errorTemplate.Execute(w, fmt.Sprintf("%v", err))
	} else {
		logsTemplate.Execute(w, logs)
	}
	footerTemplate.Execute(w, nil)
}


