package files

import (
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	now := time.Now()
	s := timeToString(now)
	now2, err := stringToTime(s)
	if err != nil {
		panic(err)
	}
	if !now.Equal(now2) {
		t.Fatalf("Expected %v, but got %v", now, now2)
	}
}