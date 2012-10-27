package web

import (
	gne "goodnewseveryone"
	"os/exec"
	"net/http"
	"fmt"
	"time"
	"sort"
	"math"
	"strconv"
	"text/template"
)

type web struct {
	gne gne.GNE
	header *template.Template
	status *template.Template
	redirectHome *template.Template
	redirectMan *template.Template
	waittime *template.Template
	invalidMinutes *template.Template
	man *template.Template
	locations *template.Template
	tasks *template.Template
	notification *template.Template
	addlocal *template.Template
	addremote *template.Template
	addtask *template.Template
	graphnodes *template.Template
	graphedges *template.Template
	diffs *template.Template
	diffLocations *template.Template
	logs *template.Template
	error *template.Template
	footer *template.Template
}

func newWeb(gne gne.GNE) *web {
	w := &web{gne: gne}
	w.header = template.Must(template.New("header").Parse(`<html><title>Good News Everyone</title>`))
	w.status = template.Must(template.New("status").Parse(`
		<div>{{if .IsRunning}}Running{{else}}Not Running{{if .IsReady}}<a href="./now">Now</a>{{end}}{{end}}</div>
		<div>{{if .IsReady}}Ready<a href="./stop">Stop</a>{{else}}Stopped<a href="./restart">Restart</a>{{end}}</div>
		<div>WaitTime {{.GetWaitTime}}<a href="./waittime">Set</a></div>
		<div><a href="./man">Task Management</a></div>
		<div><a href="./diffs">Diffs</a></div>
		<div><a href=".">Current Status</a></div>
		`))
	w.diffLocations = template.Must(template.New("diffLocations").Parse(`
		<div><a href="../">Back To Current Status</a></div>
		<table>
		{{range .}}
			<tr><td><a href="./diffs?location={{.Id}}">{{.}}</a></td></tr>
		{{end}}
		</table>
		`))
	w.redirectHome = template.Must(template.New("redirectHome").Parse(`
		<head><meta http-equiv="Refresh" content="{{.Delay}};url=../?min={{.Min}}&max={{.Max}}"></head>
		`))
	w.redirectMan = template.Must(template.New("redirectHome").Parse(`
		<head><meta http-equiv="Refresh" content="{{.}};url=../man"></head>
		`))
	w.waittime = template.Must(template.New("waittime").Parse(`
		<a href="../">Back</a>
		<form action="./waittime" method="get">
			<div>Wait Time</div>
			<input type="number" name="minutes" value="{{.}}"/> minutes
			<input type="submit" name="submit" value="set"/>
		</form>`))
	w.invalidMinutes = template.Must(template.New("invalidMinutes").Parse(`
		<div>invalid minutes received {{.}}</div>`))
	w.man = template.Must(template.New("man").Parse(`
		<div>Management</div>
		<div><a href="../">Back</a></div>
		<div><a href="./addlocal">Add Local Location</a></div>
		<div><a href="./addremote">Add Remote Location</a></div>
		<div><a href="./addtask">Add Task</a></div>
		`))
	w.locations = template.Must(template.New("locations").Parse(`
		<div>Locations</div>
		<table>
		{{range .}}
		<tr><td><div>{{.Id}}</div></td><td><a href="./removelocation?location={{.Id}}">Remove</a></td></tr>
		{{end}}
		</table>
	`))
	w.tasks = template.Must(template.New("tasks").Parse(`
		<div>Tasks</div>
		<table>
		<tr><td>Task</td><td></td><td>Last Completed Time</td></tr>
		{{range .}}
		<tr><td>{{.Id}}</td><td><a href="./removetask?task={{.Id}}">Remove</a></td><td>{{.LastCompleted}}</td></tr>
		{{end}}
		</table>
		`))
	w.notification = template.Must(template.New("notification").Parse(`
		<div>{{.}}</div>
		`))
	w.error = template.Must(template.New("error").Parse(`
		<div>An error occured: {{.}}</div>
		`))
	w.addlocal = template.Must(template.New("addlocal").Parse(`
		<div><a href="../man">Back</a></div>
		<form action="./addlocalcall" method="get">
			<div>Add Local Location</div>
			Folder<input type="text" name="local" value=""/>
			<input type="submit" name="submit" value="AddLocal"/>
		</form>
		`))
	w.addremote = template.Must(template.New("addremote").Parse(`
		<div><a href="../man">Back</a></div>
		<form action="./addremotecall" method="get">
			<div>Add Remote Location</div>
			<table>
			<tr><td>Type</td>
			<td><select name="typ"> 
        		<option value="smb" selected="selected">Samba</option>
        		<option value="ftp">FTP</option>
    		</select></td></tr>
			<tr><td>IP Address</td><td><input type="text" name="ipaddress" value=""/></td></tr>
			<tr><td>Mac</td><td><input type="text" name="mac" value=""/></td></tr>
			<tr><td>Username</td><td><input type="text" name="username" value=""/></td></tr>
			<tr><td>Password</td><td><input type="text" name="password" value=""/></td></tr>
			<tr><td>Remote Folder</td><td><input type="text" name="remote" value=""/></td></tr>
			<tr><td>Local Mounted Folder</td><td><input type="text" name="local" value=""/></td></tr>
			<tr><td><input type="submit" name="submit" value="AddRemote"/></td><td></td></tr>
			</table>
		</form>
		`))
	w.addtask = template.Must(template.New("addtask").Parse(`
		<div><a href="../man">Back</a></div>
		<form action="./addtaskcall" method="get">
			<table>
				<tr>
					<td>
						Source
					</td>
					<td>
						<select name="src">
						{{range .}}
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
						<option value="sync" selected="selected">Sync</option>
						<option value="backup" selected="selected">Backup</option>
						<option value="move" selected="selected">Move</option>
						</select>
					</td>
				</tr>
				<tr>
					<td>
						Destination
					</td>
					<td>
						<select name="dst">
						{{range .}}
						<option value="{{.Id}}" selected="selected">{{.Id}}</option>
						{{end}}
					</td>
				</tr>
				<tr><td><input type="submit" name="submit" value="AddTask"/></td><td></td></tr>
			</table>
		</form>
		`))
	w.graphnodes = template.Must(template.New("graphnodes").Parse(`
		digraph {
			{{range .}}
			"{{.Id}}";
			{{end}}
		`))
	w.graphedges = template.Must(template.New("graphedges").Parse(`
			{{range .}}
			"{{.Src}}" -> "{{.Dst}}" [label="{{.Type}}"];
			{{end}}
		}
		`))
	w.diffs = template.Must(template.New("diffs").Parse(`
		<table>
		<tr><td>Viewing Logs</td><td>{{.CurrentMin}} - {{.CurrentMax}}</td>
		<tr><td><a href="./diffs?min={{.PreviousMin}}&max={{.PreviousMax}}">Previous</a></td>
		<td><a href="./diffs?min={{.NextMin}}&max={{.NextMax}}">Next</a></td></tr>
		{{range .Contents}}
			<tr><td></td><td></td></tr>
			<tr><td>{{.At}}</td><td></td></tr>
			<tr><td></td><td></td></tr>
			{{range .Created}}
				<tr><td>+</td><td>{{.}}</td></tr>
			{{end}}
			{{range .Deleted}}
				<tr><td>-</td><td>{{.}}</td></tr>
			{{end}}
		{{end}}
		</table>
		`))
	w.logs = template.Must(template.New("logs").Parse(`
		<table>
		<tr><td>Viewing Logs</td><td>{{.CurrentMin}} - {{.CurrentMax}}</td>
		<tr><td><a href="./?min={{.PreviousMin}}&max={{.PreviousMax}}">Previous</a></td>
		<td><a href="./?min={{.NextMin}}&max={{.NextMax}}">Next</a></td></tr>
		{{range .Contents}}
			<tr><td></td><td></td></tr>
			<tr><td>{{.At}}</td><td></td></tr>
			<tr><td></td><td></td></tr>
			{{range .Lines}}
				<tr><td>{{.At.String}}</td><td>{{.Line}}</td></tr>
			{{end}}
		{{end}}
		</table>
		`))
	w.footer = template.Must(template.New("footer").Parse(`</html>`))
	return w
}

func Serve(gne gne.GNE) {
	this := newWeb(gne)
	http.HandleFunc("/restart", func(w http.ResponseWriter, r *http.Request) {
		this.handleRestart(w,r)
	})
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		this.handleStop(w,r)
	})
	http.HandleFunc("/now", func(w http.ResponseWriter, r *http.Request) {
		this.handleNow(w,r)
	})
	http.HandleFunc("/waittime", func(w http.ResponseWriter, r *http.Request) {
		this.handleWaittime(w,r)
	})
	http.HandleFunc("/man", func(w http.ResponseWriter, r *http.Request) {
		this.handleMan(w,r)
	})
	http.HandleFunc("/removelocation", func(w http.ResponseWriter, r *http.Request) {
		this.handleRemoveLocation(w,r)
	})
	http.HandleFunc("/removetask", func(w http.ResponseWriter, r *http.Request) {
		this.handleRemoveTask(w,r)
	})
	http.HandleFunc("/addlocal", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddLocal(w,r)
	})
	http.HandleFunc("/addlocalcall", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddLocalCall(w,r)
	})
	http.HandleFunc("/addremote", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddRemote(w,r)
	})
	http.HandleFunc("/addremotecall", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddRemoteCall(w,r)
	})
	http.HandleFunc("/addtask", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddTask(w,r)
	})
	http.HandleFunc("/addtaskcall", func(w http.ResponseWriter, r *http.Request) {
		this.handleAddTaskCall(w,r)
	})
	http.HandleFunc("/diffs", func(w http.ResponseWriter, r *http.Request) {
		this.handleDiffs(w, r)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		this.handleStatus(w,r)
	})
    http.ListenAndServe(":8080", nil)
}

var (
	quick = 0
	slow = 5
)

func (this *web) handleRestart(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	this.gne.Restart()
	this.redirectHome.Execute(w, quickHome)
	this.footer.Execute(w, nil)
}

func (this *web) handleStop(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	this.gne.Stop()
	this.redirectHome.Execute(w, quickHome)
	this.footer.Execute(w, nil)
}

func (this *web) handleNow(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	this.gne.Now()
	this.redirectHome.Execute(w, quickHome)
	this.footer.Execute(w, nil)
}

func (this *web) handleWaittime(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	minutes := r.FormValue("minutes")
	if len(minutes) > 0 {
		i, err := strconv.Atoi(minutes)
		if err != nil {
			this.invalidMinutes.Execute(w, minutes)
		} else {
			this.gne.SetWaitTime(time.Duration(i)*time.Minute)
		}
	}
	currentMinutes := int(this.gne.GetWaitTime() / time.Minute)
	this.waittime.Execute(w, currentMinutes)
	this.footer.Execute(w, nil)
}

func (this *web) handleMan(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	this.man.Execute(w, nil)
	this.locations.Execute(w, this.gne.GetLocations())
	this.tasks.Execute(w, this.gne.GetTasks())
	c := exec.Command("dot", "-Tsvg")
	in, err := c.StdinPipe()
	if err != nil {
		this.error.Execute(w, fmt.Sprintf("%v", err))
	} else {
		go func() { 
		this.graphnodes.Execute(in, this.gne.GetLocations())
		this.graphedges.Execute(in, this.gne.GetTasks())
		in.Close()
		}()
		data, err := c.CombinedOutput()
		if err != nil {
			this.error.Execute(w, fmt.Sprintf("%v", err))
		} else {
			fmt.Fprintf(w, "%v", string(data))
		}
	}
	this.footer.Execute(w, nil)
}

func (this *web) handleRemoveLocation(w http.ResponseWriter, r *http.Request) {
	locationId := r.FormValue("location")
	locations := this.gne.GetLocations()
	location, ok := locations[gne.LocationId(locationId)]
	if !ok {
		this.header.Execute(w, nil)
		this.redirectMan.Execute(w, slow)
		this.error.Execute(w, "location does not exist")
		this.footer.Execute(w, nil)
		return
	}
	err := this.gne.RemoveLocation(location.Id())
	if err != nil {
		this.header.Execute(w, nil)
		this.redirectMan.Execute(w, slow)
		this.error.Execute(w, fmt.Sprintf("unable to remove location: %v", err))
		this.footer.Execute(w, nil)
	} else {
		this.redirectMan.Execute(w, quick)
	}
}

func (this *web) handleRemoveTask(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	taskId := r.FormValue("task")
	tasks := this.gne.GetTasks()
	task, ok := tasks[gne.TaskId(taskId)]
	if !ok {
		this.redirectMan.Execute(w, slow)
		this.error.Execute(w, "task does not exist")
	} else {
		err := this.gne.RemoveTask(task.Id())
		if err != nil {
			this.redirectMan.Execute(w, slow)
			this.error.Execute(w, fmt.Sprintf("unable to remove task: %v", err))
		} else {
			this.redirectMan.Execute(w, quick)
		}
	}
	this.footer.Execute(w, nil)
}

func (this *web) handleAddLocal(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	this.addlocal.Execute(w, nil)
	this.footer.Execute(w, nil)
}

func (this *web) handleAddLocalCall(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	local := r.FormValue("local")
	location := gne.NewLocalLocation(local)
	err := this.gne.AddLocation(location)
	if err != nil {
		this.redirectMan.Execute(w, slow)
		this.error.Execute(w, fmt.Sprintf("unable to add local location: %v", err))
	} else {
		this.redirectMan.Execute(w, quick)
	}
	this.footer.Execute(w, nil)
}

func (this *web) handleAddRemote(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	this.addremote.Execute(w, nil)
	this.footer.Execute(w, nil)
}

func (this *web) handleAddRemoteCall(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	typ := r.FormValue("typ")
	ipaddress := r.FormValue("ipaddress")
	mac := r.FormValue("mac")
	username := r.FormValue("username")
	password := r.FormValue("password")
	remote := r.FormValue("remote")
	local := r.FormValue("local")
	location := gne.NewRemoteLocation(gne.RemoteLocationType(typ), ipaddress, mac, username, password, remote, local)
	err := this.gne.AddLocation(location)
	if err != nil {
		this.redirectMan.Execute(w, slow)
		this.error.Execute(w, fmt.Sprintf("unable to add remote location: %v", err))
	} else {
		this.redirectMan.Execute(w, quick)
	}
	this.footer.Execute(w, nil)
}

func (this *web) handleAddTask(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	this.addtask.Execute(w, this.gne.GetLocations())
	this.footer.Execute(w, nil)
}

func (this *web) handleAddTaskCall(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	typ := r.FormValue("typ")
	src := r.FormValue("src")
	dst := r.FormValue("dst")
	task := gne.NewTask(gne.LocationId(src), gne.TaskType(typ), gne.LocationId(dst))
	err := this.gne.AddTask(task)
	if err != nil {
		this.redirectMan.Execute(w, slow)
		this.error.Execute(w, fmt.Sprintf("unable to add task: %v", err))
	} else {
		this.redirectMan.Execute(w, quick)
	}
	this.footer.Execute(w, nil)
}

type timeRange struct {
	min time.Time
	max time.Time
}

func newTimeRange(minTime, maxTime string) (*timeRange, error) {
	min := time.Unix(0, 0)
	var err error = nil
	if len(minTime) > 0 {
    	min, err = time.Parse(gne.DefaultTimeFormat, minTime)
    	if err != nil {
    		return nil, err
    	}
    }
    max := time.Unix(0, math.MaxInt64)
    if len(maxTime) > 0 {
    	max, err = time.Parse(gne.DefaultTimeFormat, maxTime)
    	if err != nil {
    		return nil, err
    	}
    }
    return &timeRange{min, max}, nil
}

func (this *timeRange) dur() time.Duration {
	return time.Duration(this.max.UnixNano() - this.min.UnixNano())
}

func (this *timeRange) format() string {
	return gne.DefaultTimeFormat
}

func (this *timeRange) PreviousMin() string {
	return this.min.Add(-1*this.dur()).Format(this.format())
}

func (this *timeRange) PreviousMax() string {
	return this.max.Add(-1*this.dur()).Format(this.format())
}

func (this *timeRange) CurrentMin() string {
	return this.min.Format(this.format())
}

func (this *timeRange) CurrentMax() string {
	return this.max.Format(this.format())
}

func (this *timeRange) NextMin() string {
	return this.min.Add(this.dur()).Format(this.format())
}

func (this *timeRange) NextMax() string {
	return this.max.Add(this.dur()).Format(this.format())
}

type DiffContent struct {
	At time.Time
	Created []string
	Deleted []string
}

type diffs struct {
	*timeRange
	Contents []*DiffContent
}

func (this *web) newDiffs(location, minTime, maxTime string) (*diffs, error) {
	diffsPerLocation, err := this.gne.GetDiffs()
	if err != nil {
		return nil, err
	}
	t, err := newTimeRange(minTime, maxTime)
	if err != nil {
		return nil, err
	}
	theDiffs := diffsPerLocation[location]
	if len(minTime) == 0 && len(theDiffs) > 10 {
    	t.min = theDiffs[10].At.Add(-1*time.Nanosecond)
    }
    if len(maxTime) == 0 && len(theDiffs) > 0 {
    	t.max = theDiffs[0].At.Add(time.Nanosecond)
    }
    contents := make([]*DiffContent, 0)
    for _, d := range theDiffs {
    	if d.At.Before(t.max) && d.At.After(t.min) {
    		created, deleted, err := d.Take()
    		if err != nil {
    			return nil, err
    		}
    		contents = append(contents, &DiffContent{
    			Created: created,
    			Deleted: deleted,
    			At: d.At,
    		})
    	}
    }
    return &diffs{
    	timeRange: t,
    	Contents: contents,
    }, nil
}

func (this *web) handleDiffs(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	location := r.FormValue("location")
	minTime := r.FormValue("min")
	maxTime := r.FormValue("max")
	this.diffLocations.Execute(w, this.gne.GetLocations())
	diffs, err := this.newDiffs(location, minTime, maxTime)
	if err != nil {
		this.error.Execute(w, fmt.Sprintf("%v", err))
	} else {
		this.diffs.Execute(w, diffs)
	}
	this.footer.Execute(w, nil)
}

type logs struct {
	*timeRange
	Contents []*gne.LogContent
}

func (this *web) newLogs(minTime, maxTime string) (*logs, error) {
	logFiles, err := this.gne.GetLogs()
	if err != nil {
		return nil, err
	}
    sort.Sort(logFiles)
    t, err := newTimeRange(minTime, maxTime)
    if err != nil {
    	return nil, err
    }
    if len(minTime) == 0 && len(logFiles) > 10 {
    	t.min = logFiles[10].At.Add(-1*time.Nanosecond)
    }
    if len(maxTime) == 0 && len(logFiles) > 0 {
    	t.max = logFiles[0].At.Add(time.Nanosecond)
    }
    contents := make([]*gne.LogContent, 0)
    for _, l := range logFiles {
    	if l.At.Before(t.max) && l.At.After(t.min) {
    		content, err := l.Open()
    		if err != nil {
    			return nil, err
    		} else {
    			contents = append(contents, content)	
    		}
    	}
    }
    return &logs{
    	timeRange: t,
    	Contents: contents,
    }, nil
}

type redirectHome struct {
	Min string
	Max string
	Delay int
}

var quickHome = &redirectHome{
	Min: "",
	Max: "",
	Delay: quick,
}

func (this *web) handleStatus(w http.ResponseWriter, r *http.Request) {
	min := r.FormValue("min")
	max := r.FormValue("max")
	this.header.Execute(w, nil)
	this.redirectHome.Execute(w, &redirectHome{
		Min: min,
		Max: max,
		Delay: slow,
	})
	this.status.Execute(w, this.gne)
	logs, err := this.newLogs(min, max)
	if err != nil {
		this.error.Execute(w, fmt.Sprintf("%v", err))
	} else {
		this.logs.Execute(w, logs)
	}
	this.footer.Execute(w, nil)
}

