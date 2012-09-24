package goodnewseveryone

import (
	"os"
	"strings"
	"fmt"
	"time"
)

const DefaultTimeFormat = "2006-01-02T15:04:05Z"

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
				str := fmt.Sprintf("%v | %v\n", time.Now().Format(DefaultTimeFormat), s)
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