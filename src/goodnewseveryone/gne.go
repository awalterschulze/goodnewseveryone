package goodnewseveryone

import (
	"fmt"
	"time"
	"sync"
)

type GNE struct {
	sync.Mutex
	Kernel *kernel
	Locations locations
	Tasks tasks
	Executor *executor
	WaitTime time.Duration
	NowChan chan time.Time
	StopChan chan time.Time
	RestartChan chan time.Time
}

func NewGNE(configLocation string) *GNE {
	startupLog, err := newLog()
	if err != nil {
		panic(err)
	}
	locations, err := newLocations(startupLog, configLocation)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Locations = %v\n", locations)
	tasks, err := newTasks(startupLog, locations, configLocation)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Tasks = %v\n", tasks)
	return &GNE{
		Kernel: newKernel(),
		Locations: locations,
		Tasks: tasks,
		Executor: newExecutor(),
		WaitTime: 5*time.Minute,
		NowChan: make(chan time.Time),
		StopChan: make(chan time.Time),
		RestartChan: make(chan time.Time),
	}
}

func Main(configLocation string) {
	gne := NewGNE(configLocation)
	go gne.Start()
	gne.Serve()
}

func (this *GNE) SetWaitTime(waitTime time.Duration) {
	this.WaitTime = waitTime
}

func (this *GNE) GetWaitTime() time.Duration {
	return this.WaitTime
}

func (this *GNE) Now() {
	this.NowChan <- time.Now()
}

func (this *GNE) Restart() {
	this.RestartChan <- time.Now()
}

func (this *GNE) Stop() {
	this.StopChan <- time.Now()
}

func (this *GNE) IsReady() bool {
	return this.Kernel.ready()
}

func (this *GNE) IsRunning() bool {
	return this.Executor.IsRunning()
}

func (this *GNE) Start() {
	waitChan := time.After(1)
	for {
		select {
		case <- waitChan:
			if this.Kernel.ready() {
				this.Executor.All(this.Kernel, this.Locations, this.Tasks)
			}
		case <- this.NowChan:
			this.Executor.All(this.Kernel, this.Locations, this.Tasks)
		case <- this.StopChan:
			stopLog, err := newLog()
			if err != nil {
				panic(err)
			}
			this.Kernel.stop(stopLog)
		case <- this.RestartChan:
			this.Kernel.restart()
		}
		waitChan = time.After(this.WaitTime)
	}
}
