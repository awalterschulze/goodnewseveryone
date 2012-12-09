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
	"fmt"
	"os/exec"
)

func init() {
	http.HandleFunc("/man", func(w http.ResponseWriter, r *http.Request) {
		this.handleMan(w,r)
	})
}

var (
	manTemplate = template.Must(template.New("man").Parse(`
		<div>Management</div>
		<div><a href="../">Back</a></div>
		<div><a href="./addlocal">Add Local Location</a></div>
		<div><a href="./addremote">Add Remote Location</a></div>
		<div><a href="./addtask">Add Task</a></div>
	`))
	graphnodesTemplate = template.Must(template.New("graphnodes").Parse(`
		digraph {
			{{range .}}
			"{{.Id}}";
			{{end}}
	`))
	graphedgesTemplate = template.Must(template.New("graphedges").Parse(`
			{{range .}}
			"{{.Src}}" -> "{{.Dst}}" [label="{{.Type}}"];
			{{end}}
		}
	`))
)

func (this *web) handleMan(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	manTemplate.Execute(w, nil)
	locationsTemplate.Execute(w, this.gne.GetLocations())
	tasksTemplate.Execute(w, this.gne.GetTasks())
	c := exec.Command("dot", "-Tsvg")
	in, err := c.StdinPipe()
	if err != nil {
		errorTemplate.Execute(w, fmt.Sprintf("%v", err))
	} else {
		go func() { 
		graphnodesTemplate.Execute(in, this.gne.GetLocations())
		graphedgesTemplate.Execute(in, this.gne.GetTasks())
		in.Close()
		}()
		data, err := c.CombinedOutput()
		if err != nil {
			errorTemplate.Execute(w, fmt.Sprintf("%v", err))
		} else {
			fmt.Fprintf(w, "%v", string(data))
		}
	}
	footerTemplate.Execute(w, nil)
}
