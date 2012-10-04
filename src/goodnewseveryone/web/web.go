package web

import (
	gne "goodnewseveryone"
	"net/http"
	"fmt"
	"time"
	"strings"
	"io"
	"io/ioutil"
	"sort"
	"math"
	"strconv"
	"text/template"
)

type web struct {
	gne gne.GNE
	header *template.Template
	refresh *template.Template
	status *template.Template
	redirectHome *template.Template
	waittime *template.Template
	invalidMinutes *template.Template
	footer *template.Template
}

func newWeb(gne gne.GNE) *web {
	w := &web{gne: gne}
	w.header = template.Must(template.New("header").Parse(`<html>`))
	w.refresh = template.Must(template.New("refresh").Parse(
		`<head><meta http-equiv="Refresh" content="{{.}};url=."></head>
	`))
	w.status = template.Must(template.New("status").Parse(`
		<div>{{if .IsRunning}}Running{{else}}Not Running{{if .IsReady}}<a href="./now">Now</a>{{end}}{{end}}</div>
		<div>{{if .IsReady}}Ready<a href="./stop">Stop</a>{{else}}Stopped<a href="./restart">Restart</a>{{end}}</div>
		<div>WaitTime {{.GetWaitTime}}<a href="./waittime">Set</a></div>
		`))
	w.redirectHome = template.Must(template.New("redirectHome").Parse(`<html>
		<head><meta http-equiv="Refresh" content="1;url=../"></head>
		</html>`))
	w.waittime = template.Must(template.New("waittime").Parse(`
		<a href="../">Back</a>
		<form action="./waittime" method="get">
			<div>Wait Time</div>
			<input type="number" name="minutes" value="{{.}}"/> minutes
			<input type="submit" name="submit" value="set"/>
		</form>`))
	w.invalidMinutes = template.Must(template.New("invalidMinutes").Parse(`
		<div>invalid minutes received {{.}}</div>`))
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		this.handleStatus(w,r)
	})
    http.ListenAndServe(":8080", nil)
}

func (this *web) handleRestart(w http.ResponseWriter, r *http.Request) {
	this.gne.Restart()
	this.redirectHome.Execute(w, nil)
}

func (this *web) handleStop(w http.ResponseWriter, r *http.Request) {
	this.gne.Stop()
	this.redirectHome.Execute(w, nil)
}

func (this *web) handleNow(w http.ResponseWriter, r *http.Request) {
	this.gne.Now()
	this.redirectHome.Execute(w, nil)
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

/*func (this *web) writeButtons(w http.ResponseWriter) {
	fmt.Fprintf(w, `<form action="." method="post">`)
	minutes := int(this.gne.GetWaitTime() / time.Minute)
	fmt.Fprintf(w, `<input type="number" name="wait" value="%v" /> minutes`, minutes)
	fmt.Fprintf(w, `<input type="submit" name="action" value="restart"/>`)
	fmt.Fprintf(w, `<input type="submit" name="action" value="stop"/>`)
	fmt.Fprintf(w, `<input type="submit" name="action" value="now"/>`)
	fmt.Fprintf(w, `<input type="submit" name="action" value="refresh"/>`)
	fmt.Fprintf(w, `<input type="checkbox" name"log" value"true"/>`)
	fmt.Fprintf(w, `<input type="checkbox" name"diff" value"true"/>`)
	fmt.Fprintf(w, `</form>`)
	fmt.Fprintf(w, `<a href="./stop">stop</a>`)
	fmt.Fprintf(w, `<a href="./restart">restart</a>`)
}*/

func (this *web) writeLogs(w http.ResponseWriter, minTime, maxTime string) {
	fmt.Fprintf(w, "<table>")
	defer fmt.Fprintf(w, "</table>")
	logs, err := gne.NewLogFiles(".")
	if err != nil {
		fmt.Fprintf(w, "<tr><td>An error occured</td><td>%v</td></tr>", err)
		return
	}
    sort.Sort(logs)
    min := time.Unix(0, 0)
    if len(logs) > 10 {
    	min = logs[10].At
    }
    max := time.Unix(0, math.MaxInt64)	
    if len(logs) > 0 {
    	max = logs[0].At
    }
    if len(minTime) > 0 {
    	min, err = time.Parse(gne.DefaultTimeFormat, minTime)
    	if err != nil {
    		fmt.Fprintf(w, "<tr><td>An error occured</td><td>%v</td></tr>", err)
    	}
    }
    if len(maxTime) > 0 {
    	max, err = time.Parse(gne.DefaultTimeFormat, maxTime)
    	if err != nil {
    		fmt.Fprintf(w, "<tr><td>An error occured</td><td>%v</td></tr>", err)
    	}
    }
    dur := time.Duration(max.UnixNano() - min.UnixNano())
    fmt.Fprintf(w, `<tr><td>Viewing Logs</td><td>%v - %v</td>`, 
		min.Format(gne.DefaultTimeFormat), 
		max.Format(gne.DefaultTimeFormat))
    fmt.Fprintf(w, `<tr><td><a href="./?min=%v&max=%v">Previous</a></td>`, 
		min.Add(-1*dur).Format(gne.DefaultTimeFormat), 
		max.Add(-1*dur).Format(gne.DefaultTimeFormat))
    fmt.Fprintf(w, `<td><a href="./?min=%v&max=%v">Next</a></td></tr>`, 
		min.Add(dur).Format(gne.DefaultTimeFormat), 
		max.Add(dur).Format(gne.DefaultTimeFormat))
    for _, l := range logs {
    	if l.At.Before(max) && l.At.After(min) {
	  		fmt.Fprintf(w, "<tr><td></td><td></td></tr>")
	  		fmt.Fprintf(w, "<tr><td>%v</td><td></td></tr>", l.At)
	  		fmt.Fprintf(w, "<tr><td></td><td></td></tr>")
	    	err := writeLogFile(l, w)
	    	if err != nil {
	    		fmt.Fprintf(w, "<tr><td>An error occured</td><td>%v</td></tr>", err)	
	    		return
	    	}
    	}
    }
}

func (this *web) handleStatus(w http.ResponseWriter, r *http.Request) {
	min := r.FormValue("min")
	max := r.FormValue("max")
	this.header.Execute(w, nil)
	this.refresh.Execute(w, 5)
	this.status.Execute(w, this.gne)
	fmt.Fprintf(w, `<a href="./tasks">Task Management</a>`)
	this.writeLogs(w, min, max)
	this.footer.Execute(w, nil)
}

func (this *web) tasksHandler(w http.ResponseWriter, r *http.Request) {
	this.header.Execute(w, nil)
	fmt.Fprintf(w, `<a href="../">Status</a>`)
	this.footer.Execute(w, nil)
}

/*
import (
	"fmt"
	"net/http"
	"strings"
	"path/filepath"
	"os"
	"time"
	"sort"
	"io/ioutil"
	"path"
	"math"
	"io"
)

const (
	DefaultTimeFormat = "2006-01-02T15:04:05Z"
	Today = "today"
	All = "all"
	Range = "range"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<table>")
	defer fmt.Fprintf(w, "</table>")
	_, addr := path.Split(r.URL.Path)
	if len(addr) == 0 {
		addr = All
	}
	logs, err := newLogs(".")
	if err != nil {
		fmt.Fprintf(w, "<tr><td>An error occured</td><td>%v</td></tr>", err)
		return
	}
    sort.Sort(logs)
    min := time.Unix(0, 0)
    max := time.Unix(0, math.MaxInt64)
    switch addr {
    case Today:
    	now := time.Now()
    	min = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
    	max = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
    case Range:
    	min, err = time.Parse(DefaultTimeFormat, r.FormValue("min"))
    	if err != nil {
    		fmt.Fprintf(w, "<tr><td>time Parse error</td><td>%v</td></tr>", err)
    		return
    	}
    	max, err = time.Parse(DefaultTimeFormat, r.FormValue("max"))
    	if err != nil {
    		fmt.Fprintf(w, "<tr><td>time Parse error</td><td>%v</td></tr>", err)
    		return
    	}
    }
    
    if addr != All {
    	fmt.Fprintf(w, `<tr><td><a href="./range?min=%v&max=%v">Previous Day</a></td>`, 
    		min.Add(time.Hour*-24).Format(DefaultTimeFormat), 
    		max.Add(time.Hour*-24).Format(DefaultTimeFormat))
    	fmt.Fprintf(w, `<td><a href="./range?min=%v&max=%v">Next Day</a></td></tr>`, 
    		min.Add(time.Hour*24).Format(DefaultTimeFormat), 
    		max.Add(time.Hour*24).Format(DefaultTimeFormat))
    }
    for _, l := range logs {
    	if l.at.Before(max) && l.at.After(min) {
	    	err := l.Rows(w)
	    	if err != nil {
	    		fmt.Fprintf(w, "<tr><td>An error occured</td><td>%v</td></tr>", err)	
	    		return
	    	}
    	}
    }
}

*/

func writeLogFile(logFile *gne.LogFile, w io.Writer) error {
	data, err := ioutil.ReadFile(logFile.Filename)
	if err != nil {
		return err
	}
	dataStr := string(data)
	lines := strings.Split(dataStr, "\n")
	for _, line := range lines {
		cs := strings.SplitN(line, " | ", 2)
		if len(cs) == 2 {
			fmt.Fprintf(w, "<tr><td>%v</td><td>%v</td></tr>", cs[0], cs[1])
		}
	}
	return nil
}
