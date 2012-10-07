package goodnewseveryone

import (
	"os"
	"strings"
	"fmt"
	"time"
	"path/filepath"
	"io/ioutil"
)

const DefaultTimeFormat = "2006-01-02T15:04:05Z"
const logLineSep = " | "

type log struct {
	f *os.File
}

type Log interface {
	Write(str string)
	Run(name string, arg ...string)
	Error(err error)
	Output(output []byte)
	Close()
}

func newLog() (Log, error) {
	logFile, err := os.Create(fmt.Sprintf("gne-%v.log", time.Now().Format(DefaultTimeFormat)))
	if err != nil {
		return nil, err
	}
	return &log{logFile}, nil
}

func (this *log) Write(str string) {
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		ss := strings.Split(line, "\r")
		for _, s := range ss {
			if len(strings.TrimSpace(s)) > 0 {
				str := fmt.Sprintf("%v%v%v\n", time.Now().Format(DefaultTimeFormat), logLineSep, s)
				this.f.Write([]byte(str))
				fmt.Printf("%v", str)
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
	this.f.Close()
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

func NewLogFiles(root string) (LogFiles, error) {
	filenames := []string{}
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".log") {
			filenames = append(filenames, path)
		}
		return nil
	})
	res := make(LogFiles, len(filenames))
  	for i, filename := range filenames {
  		l, err := newLogFile(filename)
  		if err != nil {
  			return nil, err
  		}
  		res[i] = l
  	}
  	return res, nil
}

type LogFile struct {
	Filename string
	At time.Time
}

func newLogFile(filename string) (*LogFile, error) {
	timeStr := strings.Replace(strings.Replace(filename, "gne-", "", 1), ".log", "", 1)
	t, err := time.Parse(DefaultTimeFormat, timeStr)
	if err != nil {
		return nil, err
	}
	return &LogFile{filename, t}, nil
}

type LogLine struct {
	At time.Time
	Line string
}

type LogContent struct {
	At time.Time
	Lines []*LogLine
}

func (this *LogFile) Open() (*LogContent, error) {
	content := &LogContent{
		At: this.At,
		Lines: make([]*LogLine, 0),
	}
	data, err := ioutil.ReadFile(this.Filename)
	if err != nil {
		return nil, err
	}
	dataStr := string(data)
	lines := strings.Split(dataStr, "\n")
	for _, line := range lines {
		logLine := strings.SplitN(line, logLineSep, 2)
		if len(logLine) == 2 {
			t, err := time.Parse(DefaultTimeFormat, logLine[0])
			if err != nil {
				return nil, err
			}
			content.Lines = append(content.Lines, &LogLine{
				At: t,
				Line: logLine[1],
			})
		}
	}
	return content, nil
}
	
