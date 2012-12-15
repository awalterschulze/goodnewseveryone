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
	goodtime "goodnewseveryone/time"
	"math"
	"net/http"
	"text/template"
	"time"
)

func init() {
	http.HandleFunc("/diffs", func(w http.ResponseWriter, r *http.Request) {
		this.handleDiffs(w, r)
	})
}

var (
	diffLocationTemplate = template.Must(template.New("diffLocations").Parse(`
		<div><a href="../">Back To Current Status</a></div>
		<table>
		{{range .}}
			<tr><td><a href="./diffs?location={{.Id}}">{{.}}</a></td></tr>
		{{end}}
		</table>
	`))
	diffsTemplate = template.Must(template.New("diffs").Parse(`
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
)

type timeRange struct {
	min time.Time
	max time.Time
}

func newTimeRange(minTime, maxTime string) (*timeRange, error) {
	min := time.Unix(0, 0)
	var err error = nil
	if len(minTime) > 0 {
		min, err = goodtime.StringToNano(minTime)
		if err != nil {
			return nil, err
		}
	}
	max := time.Unix(0, math.MaxInt64)
	if len(maxTime) > 0 {
		max, err = goodtime.StringToNano(minTime)
		if err != nil {
			return nil, err
		}
	}
	return &timeRange{min, max}, nil
}

func (this *timeRange) dur() time.Duration {
	return time.Duration(this.max.UnixNano() - this.min.UnixNano())
}

func (this *timeRange) PreviousMin() string {
	return goodtime.TimeToString(this.min.Add(-1 * this.dur()))
}

func (this *timeRange) PreviousMax() string {
	return goodtime.TimeToString(this.max.Add(-1 * this.dur()))
}

func (this *timeRange) CurrentMin() string {
	return goodtime.TimeToString(this.min)
}

func (this *timeRange) CurrentMax() string {
	return goodtime.TimeToString(this.max)
}

func (this *timeRange) NextMin() string {
	return goodtime.TimeToString(this.min.Add(this.dur()))
}

func (this *timeRange) NextMax() string {
	return goodtime.TimeToString(this.max.Add(this.dur()))
}

type DiffContent struct {
	At      time.Time
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
		t.min = theDiffs[10].Current.Add(-1 * time.Nanosecond)
	}
	if len(maxTime) == 0 && len(theDiffs) > 0 {
		t.max = theDiffs[0].Current.Add(time.Nanosecond)
	}
	contents := make([]*DiffContent, 0)
	for _, d := range theDiffs {
		if d.Current.Before(t.max) && d.Current.After(t.min) {
			created, deleted, err := d.Take()
			if err != nil {
				return nil, err
			}
			contents = append(contents, &DiffContent{
				Created: created,
				Deleted: deleted,
				At:      d.Current,
			})
		}
	}
	return &diffs{
		timeRange: t,
		Contents:  contents,
	}, nil
}

func (this *web) handleDiffs(w http.ResponseWriter, r *http.Request) {
	location := r.FormValue("location")
	minTime := r.FormValue("min")
	maxTime := r.FormValue("max")
	execute(headerTemplate, w, nil)
	execute(diffLocationTemplate, w, this.gne.GetLocations())
	diffs, err := this.newDiffs(location, minTime, maxTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		execute(diffsTemplate, w, diffs)
	}
	execute(footerTemplate, w, nil)
}
