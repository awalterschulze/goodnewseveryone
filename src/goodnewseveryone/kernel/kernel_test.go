package kernel

import (
	"testing"
	"goodnewseveryone/log"
)

type sleep struct {
	wait chan bool
}

func newSleep() *sleep {
	return &sleep{make(chan bool)}
}

func (this *sleep) Run(log log.Log) ([]byte, error) {
	<-this.wait
	return nil, nil
}

func (this *sleep) Stop(log log.Log) {
	this.wait <- true
}

func TestSleep(t *testing.T) {
	s := newSleep()
	w := make(chan bool)
	go func() {
		if _, err := s.Run(nil); err != nil {
			panic(err)
		}
		close(w)
	}()
	s.Stop(nil)
	<-w
}

/*type Kernel interface {
	Blocked() bool
	StopAndBlock(log log.Log)
	Unblock()
	Run(log log.Log, command command.Command) (string, error)
	//run even if blocked, used for umounts
	SudoRun(log log.Log, command command.Command)
}*/

func TestRun(t *testing.T) {
	s := newSleep()
	k := NewKernel()
	w := make(chan bool)
	go func() { 
		if _, err := k.Run(nil, s); err != nil {
			panic(err)
		}
		close(w)
	}()
	s.Stop(nil)
	<-w
}

func TestStop(t *testing.T) {
	s := newSleep()
	k := NewKernel()
	w := make(chan bool)
	go func() { 
		if _, err := k.Run(nil, s); err == nil {
			t.Fatalf("Expected Error")
		}
		close(w)
	}()
	k.StopAndBlock(nil)
	<-w
}

func TestUnblock(t *testing.T) {
	k := NewKernel()
	k.StopAndBlock(nil)
	if k.Blocked() == nil {
		t.Fatalf("Expected Blocked")
	}
	k.Unblock()
	if k.Blocked() != nil {
		t.Fatalf("Expected Unblocked")
	}
}

func TestBlock(t *testing.T) {
	s := newSleep()
	k := NewKernel()
	k.StopAndBlock(nil)
	if _, err := k.Run(nil, s); err == nil {
		t.Fatalf("Expected Error")
	}
}

func TestSudoRun(t *testing.T) {
	s := newSleep()
	k := NewKernel()
	k.StopAndBlock(nil)
	w := make(chan bool)
	go func() { 
		k.SudoRun(nil, s)
		close(w)
	}()
	s.Stop(nil)
	<-w
}

func TestRunBlockSudoRun(t *testing.T) {
	s := newSleep()
	k := NewKernel()
	w := make(chan bool)
	go func() { 
		if _, err := k.Run(nil, s); err == nil {
			t.Fatalf("Expected Error")
		}
		close(w)
	}()
	k.StopAndBlock(nil)
	<-w
	s2 := newSleep()
	w2 := make(chan bool)
	go func() { 
		k.SudoRun(nil, s2)
		close(w2)
	}()
	s2.Stop(nil)
	<-w2
}

func TestManyRuns(t *testing.T) {
	num := 10
	k := NewKernel()
	wait := make(chan int)
	commands := make([]*sleep, num)
	for i := 0; i < num; i++ {
		commands[i] = newSleep()
	}
	for i := 0; i < num; i++ {
		j := i
		go func() {
			if _, err := k.Run(nil, commands[j]); err != nil {
				panic(err)
			}
			wait <- j
		}()
	}
	for i := 0; i < num; i++ {
		commands[i].Stop(nil)
		w := <-wait
		if i != w {
			t.Fatalf("Expected %v", i)
		}
	}
}


