//Copyright 2012 Walter Schulze
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package time

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const defaultTimeFormat = time.RFC3339

func NanoToString(t time.Time) string {
	return fmt.Sprintf("%v.%v", t.Format(defaultTimeFormat), t.Nanosecond())
}

func StringToNano(s string) (time.Time, error) {
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

func TimeToString(t time.Time) string {
	return t.Format(defaultTimeFormat)
}

func StringToTime(s string) (time.Time, error) {
	return time.Parse(defaultTimeFormat, s)
}
