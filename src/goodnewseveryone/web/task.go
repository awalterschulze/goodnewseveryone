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
	"goodnewseveryone/task"
	"goodnewseveryone/location"
)

func init() {
	http.HandleFunc("/removetask", func(w http.ResponseWriter, r *http.Request) {
		this.handleRemoveTask(w,r)
	})
	http.HandleFunc("/addtask", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddTask(w,r)
	})
	http.HandleFunc("/addtaskcall", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddTaskCall(w,r)
	})
}

var (
	addtaskTemplate = template.Must(template.New("addtask").Parse(`
		<div><a href="../man">Back</a></div>
		<form action="./addtaskcall" method="get">
			<table>
				<tr>
					<td>
						Name
					</td>
					<td>
						<input type="text" name="name" value=""/>
					</td>
				</tr>
				<tr>
					<td>
						Source
					</td>
					<td>
						<select name="src">
						{{range .Locations}}
						<option value="{{.Id}}">{{.Id}}</option>
						{{end}}
					</td>
				</tr>
				<tr>
					<td>
						Type
					</td>
					<td>
						<select name="typ">
						{{range .TaskTypes}}
						<option value="{{.Name}}" selected="selected">{{.Name}}</option>
						{{end}}
						</select>
					</td>
				</tr>
				<tr>
					<td>
						Destination
					</td>
					<td>
						<select name="dst">
						{{range .Locations}}
						<option value="{{.Id}}" selected="selected">{{.Id}}</option>
						{{end}}
					</td>
				</tr>
				<tr><td><input type="submit" name="submit" value="AddTask"/></td><td></td></tr>
			</table>
		</form>
	`))
	tasksTemplate = template.Must(template.New("tasks").Parse(`
		<div>Tasks</div>
		<table>
		<tr><td>Task</td><td></td><td>Last Completed Time</td></tr>
		{{range .}}
		<tr><td>{{.Id}}</td><td><a href="./removetask?task={{.Id}}">Remove</a></td><td>{{.LastCompleted}}</td></tr>
		{{end}}
		</table>
	`))
)

type taskSetup struct {
	Locations location.Locations
	TaskTypes []task.TaskType
}

func (this *web) handleAddTask(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	taskTypes, err := this.gne.GetTaskTypes()
	if err != nil {
		redirectManTemplate.Execute(w, slow)
		errorTemplate.Execute(w, fmt.Sprintf("unable to create task: %v", err))
	} else {
		setup := &taskSetup{
			Locations: this.gne.GetLocations(),
			TaskTypes: taskTypes,
		}
		addtaskTemplate.Execute(w, setup)
	}
	footerTemplate.Execute(w, nil)
}

func (this *web) handleAddTaskCall(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	name := r.FormValue("name")
	typ := r.FormValue("typ")
	src := r.FormValue("src")
	dst := r.FormValue("dst")
	taskTypes, err := this.gne.GetTaskTypes()
	if err != nil {
		redirectManTemplate.Execute(w, slow)
		errorTemplate.Execute(w, fmt.Sprintf("unable to add task: %v", err))
		footerTemplate.Execute(w, nil)
		return
	}
	var taskType task.TaskType = nil
	for i, tt := range taskTypes {
		if tt.Name() == typ {
			taskType = taskTypes[i]
		}
	}
	t := task.NewTask(name, taskType, src, dst)
	if err := this.gne.AddTask(t); err != nil {
		redirectManTemplate.Execute(w, slow)
		errorTemplate.Execute(w, fmt.Sprintf("unable to add task: %v", err))
	} else {
		redirectManTemplate.Execute(w, quick)
	}
	footerTemplate.Execute(w, nil)
}

