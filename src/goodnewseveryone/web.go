package goodnewseveryone

import (
	"net/http"
	"fmt"
	"time"
	"strings"
	"io"
	"io/ioutil"
	"path/filepath"
	"os"
	"sort"
	"math"
	"strconv"
)

func (this *GNE) Serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		this.handler(w,r)
	})
    http.ListenAndServe(":8080", nil)
}

func (this *GNE) writeHeader(w http.ResponseWriter) {
	fmt.Fprintf(w, "<html>")
}

func (this *GNE) handleAction(w http.ResponseWriter, action string, waittime string) {
	switch action {
	case "now":
		this.Now()
	case "restart":
		this.Restart()
	case "stop":
		this.Stop()
	}
	if len(action) > 0 {
		fmt.Fprintf(w, "<div>action received %v</div>", action)
	}
	i, err := strconv.Atoi(waittime)
	if err != nil {
		fmt.Fprintf(w, "<div>invalid waittime received %v</div>", waittime)
		return
	}
	this.SetWaitTime(time.Duration(i)*time.Minute)
}

func (this *GNE) writeStatus(w http.ResponseWriter) {
	if this.IsRunning() {
		fmt.Fprintf(w, "<div>Tasks are Currently Executing</div>")
	} else {
		fmt.Fprintf(w, "<div>Not Running</div>")
	}
	if !this.IsReady() {
		fmt.Fprintf(w, "<div>Stopped</div>")
	} else {
		fmt.Fprintf(w, "<div>Not Stopped</div>")
	}
}

func (this *GNE) writeButtons(w http.ResponseWriter) {
	fmt.Fprintf(w, `<form action="." method="post">`)
	minutes := int(this.GetWaitTime() / time.Minute)
	fmt.Fprintf(w, `<input type="number" name="wait" value="%v" /> minutes`, minutes)
	fmt.Fprintf(w, `<input type="submit" name="action" value="restart"/>`)
	fmt.Fprintf(w, `<input type="submit" name="action" value="stop"/>`)
	fmt.Fprintf(w, `<input type="submit" name="action" value="now"/>`)
	fmt.Fprintf(w, `<input type="submit" name="action" value="refresh"/>`)
	fmt.Fprintf(w, `<input type="checkbox" name"log" value"true"/>`)
	fmt.Fprintf(w, `<input type="checkbox" name"diff" value"true"/>`)
	fmt.Fprintf(w, `</form>`)
}

func (this *GNE) writeFooter(w http.ResponseWriter) {
	fmt.Fprintf(w, "</html>")
}

func (this *GNE) handler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	waittime := r.FormValue("wait")
	min := r.FormValue("min")
	max := r.FormValue("max")
	//logFlag := r.FormValue("log")
	this.writeHeader(w)
	this.handleAction(w, action, waittime)
	this.writeStatus(w)
	this.writeButtons(w)
	this.writeLogs(w, min, max)
	this.writeFooter(w)
}

func (this *GNE) writeLogs(w http.ResponseWriter, minTime, maxTime string) {
	fmt.Fprintf(w, "<table>")
	defer fmt.Fprintf(w, "</table>")
	logs, err := newLogFiles(".")
	if err != nil {
		fmt.Fprintf(w, "<tr><td>An error occured</td><td>%v</td></tr>", err)
		return
	}
    sort.Sort(logs)
    min := time.Unix(0, 0)
    if len(logs) > 10 {
    	min = logs[10].at
    }
    max := time.Unix(0, math.MaxInt64)	
    if len(logs) > 0 {
    	max = logs[0].at
    }
    if len(minTime) > 0 {
    	min, err = time.Parse(DefaultTimeFormat, minTime)
    	if err != nil {
    		fmt.Fprintf(w, "<tr><td>An error occured</td><td>%v</td></tr>", err)
    	}
    }
    if len(maxTime) > 0 {
    	max, err = time.Parse(DefaultTimeFormat, maxTime)
    	if err != nil {
    		fmt.Fprintf(w, "<tr><td>An error occured</td><td>%v</td></tr>", err)
    	}
    }
    dur := time.Duration(max.UnixNano() - min.UnixNano())
    fmt.Fprintf(w, `<tr><td>Viewing Logs</td><td>%v - %v</td>`, 
		min.Format(DefaultTimeFormat), 
		max.Format(DefaultTimeFormat))
    fmt.Fprintf(w, `<tr><td><a href="./?min=%v&max=%v">Previous</a></td>`, 
		min.Add(-1*dur).Format(DefaultTimeFormat), 
		max.Add(-1*dur).Format(DefaultTimeFormat))
    fmt.Fprintf(w, `<td><a href="./?min=%v&max=%v">Next</a></td></tr>`, 
		min.Add(dur).Format(DefaultTimeFormat), 
		max.Add(dur).Format(DefaultTimeFormat))
    for _, l := range logs {
    	if l.at.Before(max) && l.at.After(min) {
	  		fmt.Fprintf(w, "<tr><td></td><td></td></tr>")
	  		fmt.Fprintf(w, "<tr><td>%v</td><td></td></tr>", l.at)
	  		fmt.Fprintf(w, "<tr><td></td><td></td></tr>")
	    	err := l.Rows(w)
	    	if err != nil {
	    		fmt.Fprintf(w, "<tr><td>An error occured</td><td>%v</td></tr>", err)	
	    		return
	    	}
    	}
    }
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

type logFiles []*logFile

func (this logFiles) Len() int {
	return len(this)
}

func (this logFiles) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this logFiles) Less(i, j int) bool {
	return this[i].at.After(this[j].at)
}

type logFile struct {
	filename string
	at time.Time
}

func newLogFile(filename string) (*logFile, error) {
	timeStr := strings.Replace(strings.Replace(filename, "gne-", "", 1), ".log", "", 1)
	t, err := time.Parse(DefaultTimeFormat, timeStr)
	if err != nil {
		return nil, err
	}
	return &logFile{filename, t}, nil
}

func (this *logFile) Rows(w io.Writer) error {
	data, err := ioutil.ReadFile(this.filename)
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
	return err
}

func newLogFiles(root string) (logFiles, error) {
	filenames := []string{}
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".log") {
			filenames = append(filenames, path)
		}
		return nil
	})
	res := make(logFiles, len(filenames))
  	for i, filename := range filenames {
  		l, err := newLogFile(filename)
  		if err != nil {
  			return nil, err
  		}
  		res[i] = l
  	}
  	return res, nil
}
