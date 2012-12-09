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
	"goodnewseveryone/location"
)

func init() {
	http.HandleFunc("/addlocal", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddLocal(w,r)
	})
	http.HandleFunc("/addlocalcall", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddLocalCall(w,r)
	})
	http.HandleFunc("/removelocation", func(w http.ResponseWriter, r *http.Request) {
		this.handleRemoveLocation(w,r)
	})
	http.HandleFunc("/addremote", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddRemote(w,r)
	})
	http.HandleFunc("/addremotecall", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddRemoteCall(w,r)
	})
}

var (
	addlocalTemplate = template.Must(template.New("addlocal").Parse(`
		<div><a href="../man">Back</a></div>
		<form action="./addlocalcall" method="get">
			<div>Add Local Location</div>
			Folder<input type="text" name="local" value=""/>
			<input type="submit" name="submit" value="AddLocal"/>
		</form>
	`))
	addremoteTemplate = template.Must(template.New("addremote").Parse(`
		<div><a href="../man">Back</a></div>
		<form action="./addremotecall" method="get">
			<div>Add Remote Location</div>
			<table>
			<tr><td>Type</td>
			<td><select name="typ"> 
				{{range .}}
        		<option value="{{.Name}}">{{.Name}}</option>
        		{{end}}
    		</select></td></tr>
			<tr><td>IP Address</td><td><input type="text" name="ipaddress" value=""/></td></tr>
			<tr><td>Username</td><td><input type="text" name="username" value=""/></td></tr>
			<tr><td>Password</td><td><input type="password" name="password" value=""/></td></tr>
			<tr><td>Remote Folder</td><td><input type="text" name="remote" value=""/></td></tr>
			<tr><td><input type="submit" name="submit" value="AddRemote"/></td><td></td></tr>
			</table>
		</form>
	`))
	locationsTemplate = template.Must(template.New("locations").Parse(`
		<div>Locations</div>
		<table>
		{{range .}}
		<tr><td><div>{{.Id}}</div></td><td><a href="./removelocation?location={{.Id}}">Remove</a></td></tr>
		{{end}}
		</table>
	`))
)

func (this *web) handleRemoveLocation(w http.ResponseWriter, r *http.Request) {
	locName := r.FormValue("location")
	locations := this.gne.GetLocations()
	location, ok := locations[locName]
	if !ok {
		headerTemplate.Execute(w, nil)
		redirectManTemplate.Execute(w, slow)
		errorTemplate.Execute(w, "location does not exist")
		footerTemplate.Execute(w, nil)
		return
	}
	err := this.gne.RemoveLocation(location.Id())
	if err != nil {
		headerTemplate.Execute(w, nil)
		redirectManTemplate.Execute(w, slow)
		errorTemplate.Execute(w, fmt.Sprintf("unable to remove location: %v", err))
		footerTemplate.Execute(w, nil)
	} else {
		redirectManTemplate.Execute(w, quick)
	}
}

func (this *web) handleRemoveTask(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	taskName := r.FormValue("task")
	err := this.gne.RemoveTask(taskName)
	if err != nil {
		redirectManTemplate.Execute(w, slow)
		errorTemplate.Execute(w, fmt.Sprintf("unable to remove task: %v", err))
	} else {
		redirectManTemplate.Execute(w, quick)
	}
	footerTemplate.Execute(w, nil)
}

func (this *web) handleAddLocal(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	addlocalTemplate.Execute(w, nil)
	footerTemplate.Execute(w, nil)
}

func (this *web) handleAddLocalCall(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	local := r.FormValue("local")
	name := r.FormValue("name")
	location := location.NewLocalLocation(name, local)
	err := this.gne.AddLocation(location)
	if err != nil {
		redirectManTemplate.Execute(w, slow)
		errorTemplate.Execute(w, fmt.Sprintf("unable to add local location: %v", err))
	} else {
		redirectManTemplate.Execute(w, quick)
	}
	footerTemplate.Execute(w, nil)
}

func (this *web) handleAddRemote(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	types, err := this.gne.GetRemoteLocationTypes()
	if err != nil {
		redirectManTemplate.Execute(w, slow)
		errorTemplate.Execute(w, fmt.Sprintf("unable to add remote location: %v", err))
	} else {
		addremoteTemplate.Execute(w, types)
	}
	footerTemplate.Execute(w, nil)
}

func (this *web) handleAddRemoteCall(w http.ResponseWriter, r *http.Request) {
	headerTemplate.Execute(w, nil)
	name := r.FormValue("name")
	typ := r.FormValue("typ")
	ipaddress := r.FormValue("ipaddress")
	username := r.FormValue("username")
	password := r.FormValue("password")
	remote := r.FormValue("remote")
	local := r.FormValue("local")
	types, err := this.gne.GetRemoteLocationTypes()
	if err != nil {
		redirectManTemplate.Execute(w, slow)
		errorTemplate.Execute(w, fmt.Sprintf("unable to add remote location: %v", err))
		footerTemplate.Execute(w, nil)
		return
	}
	var mount, unmount string
	for i, t := range types {
		if t.Name == typ {
			mount = types[i].Mount
			unmount = types[i].Unmount
		}
	}
	location := location.NewRemoteLocation(name, typ, ipaddress, username, password, remote, local, mount, unmount)
	if err := this.gne.AddLocation(location); err != nil {
		redirectManTemplate.Execute(w, slow)
		errorTemplate.Execute(w, fmt.Sprintf("unable to add remote location: %v", err))
	} else {
		redirectManTemplate.Execute(w, quick)
	}
	footerTemplate.Execute(w, nil)
}

