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

package files

import (
	"fmt"
	"testing"
	"time"
)

func TestNewCloseDelete(t *testing.T) {
	store := NewFiles(".")
	key, _ := strToKey(keyToStr(time.Now()))
	if err := store.NewLogSession(key); err != nil {
		panic(err)
	}
	logSessions := store.ListLogSessions()
	if len(logSessions) != 1 {
		t.Fatalf("not one log session = %v", logSessions)
	}
	if !logSessions[0].Equal(key) {
		t.Fatalf("one session is not the session that was created key = %v, logSession = %v", key, logSessions[0])
	}
	if err := store.CloseLogSession(key); err != nil {
		panic(err)
	}
	logSessions = store.ListLogSessions()
	if len(logSessions) != 1 {
		t.Fatalf("not one log session = %v", logSessions)
	}
	if !logSessions[0].Equal(key) {
		t.Fatalf("one session is not the session that was created key = %v, logSession = %v", key, logSessions[0])
	}
	if err := store.DeleteLogSession(key); err != nil {
		panic(err)
	}
	logSessions = store.ListLogSessions()
	if len(logSessions) != 0 {
		t.Fatalf("expected zero log sessions = %v", logSessions)
	}
}

func TestWriteRead(t *testing.T) {
	store := NewFiles(".")
	key, _ := strToKey(keyToStr(time.Now()))
	if err := store.NewLogSession(key); err != nil {
		panic(err)
	}
	num := 100
	for i := 0; i < num; i++ {
		if err := store.WriteToLogSession(key, fmt.Sprintf("%v", i)); err != nil {
			panic(err)
		}
	}
	ts, cs, err := store.ReadFromLogSession(key)
	if err != nil {
		panic(err)
	}
	if len(ts) != num && len(cs) != num {
		t.Fatalf("times and contents are not the right length, times = %v, contents = %v", len(ts), len(cs))
	}
	if err := store.CloseLogSession(key); err != nil {
		panic(err)
	}
	ts, cs, err = store.ReadFromLogSession(key)
	if err != nil {
		panic(err)
	}
	if len(ts) != num && len(cs) != num {
		t.Fatalf("times and contents are not the right length, times = %v, contents = %v", len(ts), len(cs))
	}
	if err := store.DeleteLogSession(key); err != nil {
		panic(err)
	}
	for i := 0; i < num; i++ {
		if cs[i] != fmt.Sprintf("%v", i) {
			t.Fatalf("not in correct order %v != %v", cs[i], i)
		}
	}
}

func TestDeleteOpen(t *testing.T) {
	store := NewFiles(".")
	key, _ := strToKey(keyToStr(time.Now()))
	if err := store.NewLogSession(key); err != nil {
		panic(err)
	}
	if err := store.DeleteLogSession(key); err == nil {
		t.Fatalf("should not be able to detele open log session")
	}
	if err := store.CloseLogSession(key); err != nil {
		panic(err)
	}
	if err := store.DeleteLogSession(key); err != nil {
		panic(err)
	}
}
