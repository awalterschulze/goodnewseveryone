package goodnewseveryone

import (
	"fmt"
	"time"
	"sync"
)

type GNE interface {
	AddLocation(loc Location) error
	RemoveLocation(loc Location) error
	GetLocations() Locations
	AddTask(task Task) error
	RemoveTask(task Task) error
	GetTasks() Tasks
	SetWaitTime(waitTime time.Duration)
	GetWaitTime() time.Duration
	Now()
	Restart()
	Stop()
	IsReady() bool
	IsRunning() bool
	Start()
}

type gne struct {
	sync.Mutex
	kernel *kernel
	locations Locations
	tasks Tasks
	executor *executor
	waitTime time.Duration
	nowChan chan time.Time
	stopChan chan time.Time
	restartChan chan time.Time
}

func ConfigToGNE(configLocation string) GNE {
	startupLog, err := newLog()
	if err != nil {
		panic(err)
	}
	gne := &gne{
		kernel: newKernel(),
		executor: newExecutor(),
		waitTime: 5*time.Minute,
		nowChan: make(chan time.Time),
		stopChan: make(chan time.Time),
		restartChan: make(chan time.Time),
		tasks: make(Tasks),
		locations : make(Locations),
	}
	locations, err := configToLocations(startupLog, configLocation)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Locations = %v\n", locations)
	tasks, err := configToTasks(startupLog, configLocation)
	if err != nil {
		panic(err)
	}
	for _, l := range locations {
		err := gne.AddLocation(l)	
		if err != nil {
			startupLog.Error(err)
			panic(err)
		}
	}
	for _, t := range tasks {
		err := gne.AddTask(t)
		if err != nil {
			startupLog.Error(err)
			panic(err)
		}
	}
	fmt.Printf("Tasks = %v\n", gne.GetTasks())
	return gne
}

func (this *gne) AddLocation(loc Location) error {
	return this.locations.Add(loc)
}

func (this *gne) RemoveLocation(loc Location) error {
	return this.locations.Remove(loc)
}

func (this *gne) GetLocations() Locations {
	return this.locations
}

func (this *gne) AddTask(task Task) error {
	if _, ok := this.locations[task.Src]; !ok {
		return errUnknownLocation
	}
	if _, ok := this.locations[task.Dst]; !ok {
		return errUnknownLocation
	}
	return this.tasks.Add(task)
}

func (this *gne) RemoveTask(task Task) error {
	return this.tasks.Remove(task)
}

func (this *gne) GetTasks() Tasks {
	return this.tasks
}

func (this *gne) SetWaitTime(waitTime time.Duration) {
	this.waitTime = waitTime
}

func (this *gne) GetWaitTime() time.Duration {
	return this.waitTime
}

func (this *gne) Now() {
	this.nowChan <- time.Now()
}

func (this *gne) Restart() {
	this.restartChan <- time.Now()
}

func (this *gne) Stop() {
	this.stopChan <- time.Now()
}

func (this *gne) IsReady() bool {
	return this.kernel.ready()
}

func (this *gne) IsRunning() bool {
	return this.executor.IsRunning()
}

func (this *gne) Start() {
	waitChan := time.After(1)
	for {
		select {
		case <- waitChan:
			if this.kernel.ready() {
				this.executor.All(this.kernel, this.locations, this.tasks)
			}
		case <- this.nowChan:
			this.executor.All(this.kernel, this.locations, this.tasks)
		case <- this.stopChan:
			stopLog, err := newLog()
			if err != nil {
				panic(err)
			}
			this.kernel.stop(stopLog)
		case <- this.restartChan:
			this.kernel.restart()
		}
		waitChan = time.After(this.waitTime)
	}
}
