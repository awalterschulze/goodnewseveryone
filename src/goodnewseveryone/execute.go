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

func (this *executor) All(kernel *kernel, locations Locations, tasks Tasks) {
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

func (this *executor) all(kernel *kernel, locations Locations, tasks Tasks) {
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

func (this *executor) one(log Log, kernel *kernel, locations Locations, task Task) error {
	if !kernel.ready() {
		return errPaused
	}
	src, ok := locations[task.Src]
	if !ok {
		log.Error(errUnknownLocation)
		return errUnknownLocation
	}
	dst, ok := locations[task.Dst]
	if !ok {
		log.Error(errUnknownLocation)
		return errUnknownLocation
	}
	output, err := kernel.run(log, src.newLocateCommand())
	if err != nil {
		return err
	}
	if !src.located(log, output) {
		return nil
	}
	output, err = kernel.run(log, dst.newLocateCommand())
	if err != nil {
		return err
	}
	if !dst.located(log, output) {
		return nil
	}
	_, err = kernel.run(log, src.newMountCommand())
	if err != nil {
		return err
	}
	defer kernel.overrun(log, src.newUmountCommand())
	_, err = kernel.run(log, dst.newMountCommand())
	if err != nil {
		return err
	}
	defer kernel.overrun(log, dst.newUmountCommand())
	t, err := task.newCommand(locations)
	if err != nil {
		return err
	}
	if err := writeList(src); err != nil {
		return err
	}
	if err := writeList(dst); err != nil {
		return err
	}
	_, err = kernel.run(log, t)
	if err != nil {
		return err
	}
	if err := writeList(src); err != nil {
		return err
	}
	if err := writeList(dst); err != nil {
		return err
	}
	return nil
}
