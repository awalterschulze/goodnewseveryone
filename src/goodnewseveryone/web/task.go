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
	"goodnewseveryone/location"
	"goodnewseveryone/task"
	"net/http"
	"text/template"
)

func init() {
	http.HandleFunc("/removetask", func(w http.ResponseWriter, r *http.Request) {
		this.handleRemoveTask(w, r)
	})
	http.HandleFunc("/addtask", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddTask(w, r)
	})
	http.HandleFunc("/addtaskcall", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddTaskCall(w, r)
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
)

type taskSetup struct {
	Locations location.Locations
	TaskTypes []task.TaskType
}

func (this *web) handleRemoveTask(w http.ResponseWriter, r *http.Request) {
	taskName, err := formValue(w, r, "name")
	if err != nil {
		return
	}
	if err := this.gne.RemoveTask(taskName); err != nil {
		httpError(w, fmt.Sprintf("unable to remove task: %v", err))
		return
	}
	redirectMan(w, r)
}

func (this *web) handleAddTask(w http.ResponseWriter, r *http.Request) {
	execute(headerTemplate, w, nil)
	taskTypes, err := this.gne.GetTaskTypes()
	if err != nil {
		httpError(w, fmt.Sprintf("unable to create task: %v", err))
		return
	}
	setup := &taskSetup{
		Locations: this.gne.GetLocations(),
		TaskTypes: taskTypes,
	}
	execute(addtaskTemplate, w, setup)
	execute(footerTemplate, w, nil)
}

func (this *web) handleAddTaskCall(w http.ResponseWriter, r *http.Request) {
	name, err := formValue(w, r, "name")
	if err != nil {
		return
	}
	typ, err := formValue(w, r, "typ")
	if err != nil {
		return
	}
	src, err := formValue(w, r, "src")
	if err != nil {
		return
	}
	dst, err := formValue(w, r, "dst")
	if err != nil {
		return
	}
	taskTypes, err := this.gne.GetTaskTypes()
	if err != nil {
		httpError(w, fmt.Sprintf("unable to add task: %v", err))
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
		httpError(w, fmt.Sprintf("unable to add task: %v", err))
		return
	}
	redirectMan(w, r)
}
