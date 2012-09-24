package goodnewseveryone

import (
	"fmt"
	"time"
)

var (
	Kernel = newKernel()
	Locations = locations(nil)
	Tasks = tasks(nil)
	Executor = newExecutor()
	WaitTime = 5*time.Minute
	NowChan = make(chan time.Time)
	StopChan = make(chan time.Time)
	RestartChan = make(chan time.Time)
)

func Main(configLocation string) {
	startupLog, err := newLog()
	if err != nil {
		panic(err)
	}
	Locations, err = newLocations(startupLog, configLocation)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Locations = %v\n", Locations)
	Tasks, err = newTasks(startupLog, Locations, configLocation)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Tasks = %v\n", Tasks)
	loop()
}

func loop() {
	waitChan := time.After(time.Second)
	for {
		select {
		case <- waitChan:
			if Kernel.ready() {
				Executor.All()
			}
		case <- NowChan:
			Executor.All()
		case <- StopChan:
			stopLog, err := newLog()
			if err != nil {
				panic(err)
			}
			Kernel.stop(stopLog)
		case <- RestartChan:
			Kernel.restart()
		}
		waitChan = time.After(WaitTime)
	}
}
