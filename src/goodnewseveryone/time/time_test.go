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
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	now := time.Now()
	n := now.Nanosecond()
	now = now.Add(time.Duration(-1 * n))
	s := TimeToString(now)
	now2, err := StringToTime(s)
	if err != nil {
		panic(err)
	}
	if !now.Equal(now2) {
		t.Fatalf("Expected %v, but got %v", now, now2)
	}
}

func TestNano(t *testing.T) {
	now := time.Now()
	s := NanoToString(now)
	now2, err := StringToNano(s)
	if err != nil {
		panic(err)
	}
	if !now.Equal(now2) {
		t.Fatalf("Expected %v, but got %v", now, now2)
	}
}
