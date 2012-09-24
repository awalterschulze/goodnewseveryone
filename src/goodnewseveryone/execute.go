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

func (this *executor) All() {
	this.Lock()
	if this.running {
		return
	}
	this.running = true
	this.Unlock()
	this.all()
	this.running = false
}

func (this *executor) all() {
	log, err := newLog()
	if err != nil {
		panic(err)
	}
	for _, task := range Tasks {
		log.Write(fmt.Sprintf("Executing Task %v", task))
		err := this.one(log, task)
		if err != nil {
			log.Error(err)	
		}
		log.Write(fmt.Sprintf("Executed Task %v", task))
	}
}

func (this *executor) one(log Log, task Task) error {
	if !Kernel.ready() {
		return errPaused
	}
	src, ok := Locations[task.Src]
	if !ok {
		log.Error(errInvalidLocation)
		return errInvalidLocation
	}
	dst, ok := Locations[task.Dst]
	if !ok {
		log.Error(errInvalidLocation)
		return errInvalidLocation
	}
	output, err := Kernel.run(log, src.NewLocateCommand())
	if err != nil {
		return err
	}
	if !src.Located(log, output) {
		return nil
	}
	output, err = Kernel.run(log, dst.NewLocateCommand())
	if err != nil {
		return err
	}
	if !dst.Located(log, output) {
		return nil
	}
	_, err = Kernel.run(log, src.NewMountCommand())
	if err != nil {
		return err
	}
	defer Kernel.overrun(log, src.NewUmountCommand())
	_, err = Kernel.run(log, dst.NewMountCommand())
	if err != nil {
		return err
	}
	defer Kernel.overrun(log, dst.NewUmountCommand())
	_, err = Kernel.run(log, task.NewCommand())
	if err != nil {
		return err
	}
	return nil
}
