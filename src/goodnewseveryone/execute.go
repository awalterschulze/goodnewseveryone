package goodnewseveryone

import (
	"sync"
	"fmt"
)

type executor struct {
	sync.Mutex
	running bool
}

func newExecutor() *executor {
	return &executor{}
}

func (this *executor) IsRunning() bool {
	return this.running
}

func (this *executor) All(kernel *kernel, locations locations, tasks tasks) {
	this.Lock()
	if this.running {
		return
	}
	this.running = true
	this.Unlock()
	this.all(kernel, locations, tasks)
	this.Lock()
	this.running = false
	this.Unlock()
}

func (this *executor) all(kernel *kernel, locations locations, tasks tasks) {
	log, err := newLog()
	if err != nil {
		panic(err)
	}
	for _, task := range tasks {
		log.Write(fmt.Sprintf("Executing Task %v", task))
		err := this.one(log, kernel, locations, task)
		if err != nil {
			log.Error(err)	
		}
		log.Write(fmt.Sprintf("Executed Task %v", task))
	}
}

func (this *executor) one(log Log, kernel *kernel, locations locations, task Task) error {
	if !kernel.ready() {
		return errPaused
	}
	src, ok := locations[task.Src]
	if !ok {
		log.Error(errInvalidLocation)
		return errInvalidLocation
	}
	dst, ok := locations[task.Dst]
	if !ok {
		log.Error(errInvalidLocation)
		return errInvalidLocation
	}
	output, err := kernel.run(log, src.NewLocateCommand())
	if err != nil {
		return err
	}
	if !src.Located(log, output) {
		return nil
	}
	output, err = kernel.run(log, dst.NewLocateCommand())
	if err != nil {
		return err
	}
	if !dst.Located(log, output) {
		return nil
	}
	_, err = kernel.run(log, src.NewMountCommand())
	if err != nil {
		return err
	}
	defer kernel.overrun(log, src.NewUmountCommand())
	_, err = kernel.run(log, dst.NewMountCommand())
	if err != nil {
		return err
	}
	defer kernel.overrun(log, dst.NewUmountCommand())
	t, err := task.NewCommand(locations)
	if err != nil {
		return err
	}
	_, err = kernel.run(log, t)
	if err != nil {
		return err
	}
	return nil
}
