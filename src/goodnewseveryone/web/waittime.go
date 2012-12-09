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
	"text/template"
	"net/http"
	"strconv"
	"time"
)

func init() {
	http.HandleFunc("/waittime", func(w http.ResponseWriter, r *http.Request) {
		this.handleWaittime(w,r)
	})
}

var (
	waittimeTemplate = template.Must(template.New("waittime").Parse(`
		<a href="../">Back</a>
		<form action="./waittime" method="get">
			<div>Wait Time</div>
			<input type="number" name="minutes" value="{{.}}"/> minutes
			<input type="submit" name="submit" value="set"/>
		</form>`))
	invalidMinutesTemplate = template.Must(template.New("invalidMinutes").Parse(`
		<div>invalid minutes received {{.}}</div>`))
)

func (this *web) handleWaittime(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	minutes := r.FormValue("minutes")
	if len(minutes) > 0 {
		i, err := strconv.Atoi(minutes)
		if err != nil {
			invalidMinutesTemplate.Execute(w, minutes)
		} else {
			this.gne.SetWaitTime(time.Duration(i)*time.Minute)
		}
	}
	currentMinutes := int(this.gne.GetWaitTime() / time.Minute)
	waittimeTemplate.Execute(w, currentMinutes)
	footerTemplate.Execute(w, nil)
}