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
	"goodnewseveryone/log"
    "time"
    "sort"
)

var (
	logsTemplate = template.Must(template.New("logs").Parse(`
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
)

type logs struct {
	*timeRange
	Contents []*log.LogOpenContent
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
    contents := make([]*log.LogOpenContent, 0)
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

