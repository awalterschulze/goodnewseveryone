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

package log

import (
	"strings"
	"fmt"
	"time"
	"sort"
	"goodnewseveryone/store"
)

const DefaultTimeFormat = "2006-01-02T15:04:05Z"
const logLineSep = " | "

type log struct {
	store store.LogStore
	sessionKey time.Time
}

type Log interface {
	Write(str string)
	Run(name string, arg ...string)
	Error(err error)
	Output(output []byte)
	Close()
}

func NewLog(now time.Time, store store.LogStore) (Log, error) {
	err := store.NewLogSession(now)
	if err != nil {
		return nil, err
	}
	return &log{store, now}, nil
}

func (this *log) Write(str string) {
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		ss := strings.Split(line, "\r")
		for _, s := range ss {
			if len(strings.TrimSpace(s)) > 0 {
				this.store.WriteToLogSession(this.sessionKey, strings.TrimSpace(s))
				fmt.Printf("%v\n", strings.TrimSpace(s))
			}	
		}
	}
}

func (this *log) Run(name string, arg ...string) {
	this.Write(fmt.Sprintf("> %v %v\n", name, strings.Join(arg, " ")))
}

func (this *log) Error(err error) {
	this.Write(fmt.Sprintf("ERROR: %v", err))
}

func (this *log) Output(output []byte) {
	this.Write(fmt.Sprintf("%v\n", string(output)))
}

func (this *log) Close() {
	this.store.CloseLogSession(this.sessionKey)
}

type LogContens []*LogContent

func (this LogContens) Len() int {
	return len(this)
}

func (this LogContens) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this LogContens) Less(i, j int) bool {
	return this[i].At.After(this[j].At)
}

func NewLogContents(store store.LogStore) (LogContens, error) {
	times := store.ListLogSessions()
	res := make(LogContens, len(times))
  	for i, t := range times {
  		res[i] = newLogContent(store, t)
  	}
  	return res, nil
}

type LogContent struct {
	store store.LogStore
	At time.Time
}

func newLogContent(store store.LogStore, at time.Time) *LogContent {
	return &LogContent{store, at}
}

type LogLine struct {
	Number int
	At time.Time
	Line string
}

type LogLines []*LogLine

func (this LogLines) Len() int {
	return len(this)
}

func (this LogLines) Less(i, j int) bool {
	return this[i].Number > this[j].Number
}

func (this LogLines) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

type LogOpenContent struct {
	At time.Time
	Lines LogLines
}

func (this *LogContent) Open() (*LogOpenContent, error) {
	ts, cs, err := this.store.ReadFromLogSession(this.At)
	if err != nil {
		return nil, err
	}
	content := &LogOpenContent{
		At: this.At,
		Lines: make(LogLines, 0),
	}
	for i := 0; i < len(ts); i++ {
		content.Lines = append(content.Lines, &LogLine{
			Number: i,
			At: ts[i],
			Line: cs[i],
		})
	}
	sort.Sort(content.Lines)
	return content, nil
}
	
