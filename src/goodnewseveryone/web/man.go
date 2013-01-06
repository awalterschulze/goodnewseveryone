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
	"fmt"
	"net/http"
	"os/exec"
	"text/template"
	"time"
)

func init() {
	http.HandleFunc("/man", func(w http.ResponseWriter, r *http.Request) {
		this.handleMan(w, r)
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
			"{{.}}";
			{{end}}
	`))
	graphedgesTemplate = template.Must(template.New("graphedges").Parse(`
			{{range .}}
			"{{.Src}}" -> "{{.Dst}}" [label="{{.TaskTypeName}}"];
			{{end}}
		}
	`))
	locationsTemplate = template.Must(template.New("locations").Parse(`
		<div>Locations</div>
		<table>
		{{range .}}
		<tr><td><div>{{.Id}}</div></td><td><a href="./removelocation?name={{.Id}}">Remove</a></td></tr>
		{{end}}
		</table>
	`))
	tasksTemplate = template.Must(template.New("tasks").Parse(`
		<div>Tasks</div>
		<table>
		<tr>
			<td>Task</td>
			<td></td>
			<td>Last Completed Time</td>
			<td></td>
		</tr>
		{{range .}}
		<tr>
			<td>{{.Name}}</td>
			<td><a href="./removetask?name={{.Name}}">Remove</a></td>
			<td>{{.LastCompleted}}</td>
			<td><a href="./nowtask?name={{.Name}}">Now</a></td>
		</tr>
		{{end}}
		</table>
	`))
)

type taskItem struct {
	Name          string
	LastCompleted time.Time
	Src           string
	Dst           string
	TaskTypeName  string
}

func (this *web) handleMan(w http.ResponseWriter, r *http.Request) {
	execute(headerTemplate, w, nil)
	execute(manTemplate, w, nil)
	execute(locationsTemplate, w, this.gne.GetLocations())
	gettasks := this.gne.GetTasks()
	tasks := make([]*taskItem, 0, len(gettasks.List()))
	for _, t := range gettasks.List() {
		task := gettasks.Get(t)
		tasks = append(tasks, &taskItem{
			Name:          t,
			LastCompleted: task.LastCompleted(),
			Src:           task.Src(),
			Dst:           task.Dst(),
			TaskTypeName:  task.TaskTypeName(),
		})
	}
	execute(tasksTemplate, w, tasks)
	c := exec.Command("dot", "-Tsvg")
	in, err := c.StdinPipe()
	if err != nil {
		httpError(w, err.Error())
		return
	}
	locations := this.gne.GetLocations()
	locs := make([]string, 0, len(locations))
	for name, _ := range locations {
		locs = append(locs, name)
	}
	go func() {
		graphnodesTemplate.Execute(in, locs)
		graphedgesTemplate.Execute(in, tasks)
		in.Close()
	}()
	data, err := c.CombinedOutput()
	if err != nil {
		execute(notificationTemplate, w, err.Error())
		return
	}
	fmt.Fprintf(w, "%v", string(data))
	execute(footerTemplate, w, nil)
}
