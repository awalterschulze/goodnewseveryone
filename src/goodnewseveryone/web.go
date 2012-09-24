package goodnewseveryone

import (
	"time"
)

func SetWaitTime(dur time.Duration) {
	WaitTime = dur
}

func Now() {
	NowChan <- time.Now()
}

func Restart() {
	RestartChan <- time.Now()
}

func Stop() {
	StopChan <- time.Now()
}
