package log

import (
	"strings"
	"fmt"
	"time"
	"sort"
)

const DefaultTimeFormat = "2006-01-02T15:04:05Z"
const logLineSep = " | "

type log struct {
	store LogStore
	sessionKey time.Time
}

type LogStore interface {
	NewLogSession(key time.Time) error
	ListLogSessions() []time.Time
	ReadFromLogSession(key time.Time) ([]time.Time, []string, error)
	WriteToLogSession(key time.Time, line string) error
	DeleteLogSession(key time.Time) error
	CloseLogSession(key time.Time) error
}

type Log interface {
	Write(str string)
	Run(name string, arg ...string)
	Error(err error)
	Output(output []byte)
	Close()
}

func NewLog(now time.Time, store LogStore) (Log, error) {
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

type LogFiles []*LogFile

func (this LogFiles) Len() int {
	return len(this)
}

func (this LogFiles) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this LogFiles) Less(i, j int) bool {
	return this[i].At.After(this[j].At)
}

func NewLogFiles(store LogStore) (LogFiles, error) {
	times := store.ListLogSessions()
	res := make(LogFiles, len(times))
  	for i, t := range times {
  		res[i] = newLogFile(store, t)
  	}
  	return res, nil
}

type LogFile struct {
	store LogStore
	At time.Time
}

func newLogFile(store LogStore, at time.Time) *LogFile {
	return &LogFile{store, at}
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

type LogContent struct {
	At time.Time
	Lines LogLines
}

func (this *LogFile) Open() (*LogContent, error) {
	ts, cs, err := this.store.ReadFromLogSession(this.At)
	if err != nil {
		return nil, err
	}
	content := &LogContent{
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
	
