package goodnewseveryone

import (
	"sync"
	"errors"
)

var (
	errPaused = errors.New("Paused")
)

type kernel struct {
	sync.Mutex
	running *command
	paused bool
}

func newKernel() *kernel {
	return &kernel{}
}

func (this *kernel) ready() bool {
	this.Lock()
	defer this.Unlock()
	return !this.paused
}

func (this *kernel) restart() {
	this.Lock()
	this.paused = false
	this.Unlock()
}

func (this *kernel) stop(log Log) {
	this.Lock()
	this.paused = true
	if this.running != nil {
		this.running.stop(log)
	}
	this.Unlock()
}

func (this *kernel) run(log Log, command *command) (string, error) {
	this.Lock()
	if this.paused {
		this.Unlock()
		return "", errPaused
	}
	this.running = command
	this.Unlock()
	if this.running == nil {
		return "", nil
	}
	output, err := this.running.run(log)
	if err != nil {
		return "", err
	}
	this.Lock()
	this.running = nil
	this.Unlock()
	return string(output), err
}

func (this *kernel) overrun(log Log, command *command) {
	this.Lock()
	defer this.Unlock()
	if command == nil {
		return
	}
	this.running = command
	this.running.run(log)
	this.running = nil
}
