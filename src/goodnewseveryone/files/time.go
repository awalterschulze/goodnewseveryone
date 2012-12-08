package files

import (
	"time"
	"fmt"
	"strings"
	"strconv"
)

const defaultTimeFormat = time.RFC3339

func timeToString(t time.Time) string {
	return fmt.Sprintf("%v.%v", t.Format(defaultTimeFormat), t.Nanosecond())
}

func stringToTime(s string) (time.Time, error) {
	ss := strings.Split(s, ".")
	if len(ss) != 2 {
		return time.Parse(defaultTimeFormat, s)	
	}
	t, err := time.Parse(defaultTimeFormat, ss[0])
	if err != nil {
		fmt.Printf("time.Parse error = %v\n", err)
		return t, err
	}
	n, err := strconv.Atoi(ss[1])
	if err != nil {
		fmt.Printf("strconv.Atoi error = %v\n", err)
		return t, err
	}
	t = t.Add(time.Duration(n))
	return t, err
}
