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
)

var (
	headerTemplate = template.Must(template.New("header").Parse(`<html><title>Good News Everyone</title>`))
	notificationTemplate = template.Must(template.New("notification").Parse(`<div>{{.}}</div>`))
	errorTemplate = template.Must(template.New("error").Parse(`<div>An error occured: {{.}}</div>`))
	footerTemplate = template.Must(template.New("footer").Parse(`</html>`))
)

var (
	redirectHomeTemplate = template.Must(template.New("redirectHome").Parse(`
		<head><meta http-equiv="Refresh" content="{{.Delay}};url=../?min={{.Min}}&max={{.Max}}"></head>
	`))
	redirectManTemplate = template.Must(template.New("redirectHome").Parse(`
		<head><meta http-equiv="Refresh" content="{{.}};url=../man"></head>
	`))
)

const (
	slow = 5
)

type home struct {
	Min string
	Max string
	Delay int
}

func execute(template *template.Template, w http.ResponseWriter, data interface{}) {
	if err := template.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func redirectMan(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "../man", http.StatusOK)
}

func redirectHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "../", http.StatusOK)
}

func httpError(w http.ResponseWriter, errString string) {
	http.Error(w, errString, http.StatusInternalServerError)
}


