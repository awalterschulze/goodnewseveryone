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

package log

import (
	"testing"
	"errors"
	"time"
	"goodnewseveryone/files"
	"goodnewseveryone/store"
	"fmt"
	"sort"
)

type mockStore struct {
	key time.Time
	lines []string
	times []time.Time
}

func (this *mockStore) NewLogSession(key time.Time) error {
	this.key = key
	this.lines = make([]string, 0)
	return nil
}

func (this *mockStore) ListLogSessions() []time.Time {
	if this.lines == nil {
		return nil
	}
	return []time.Time{this.key}
}

func (this *mockStore) ReadFromLogSession(key time.Time) ([]time.Time, []string, error) {
	if key != this.key {
		panic("wrong key")
	}
	return this.times, this.lines, nil
}

func (this *mockStore) WriteToLogSession(key time.Time, line string) error {
	if key != this.key {
		panic("wrong key")
	}
	this.lines = append(this.lines, line)
	this.times = append(this.times, time.Now())
	return nil
}

func (this *mockStore) DeleteLogSession(key time.Time) error {
	if key != this.key {
		panic("wrong key")
	}
	if this.lines != nil {
		panic("session still open")
	}
	return nil
}

func (this *mockStore) CloseLogSession(key time.Time) error {
	if key != this.key {
		panic("wrong key")
	}
	this.lines = nil
	this.times = nil
	return nil
}

func testNewWriteRunErrorOutputReadClose(t *testing.T, store store.LogStore) {
	l, err := NewLog(time.Now(), store)
	if err != nil {
		panic(err)
	}
	key := l.(*log).sessionKey
	l.Write("1")
	l.Run("2", "3", "4")
	l.Error(errors.New("5"))
	l.Output([]byte{'6'})
	_, cs, err := store.ReadFromLogSession(key)
	if err != nil {
		panic(err)
	}
	if len(cs) != 4 {
		t.Fatalf("wrong number of lines = %v", cs)
	}
	l.Close()
	if err := store.DeleteLogSession(key); err != nil {
		panic(err)
	}
	if len(store.ListLogSessions()) != 0 {
		t.Fatalf("logs are not deleted")
	}
}

func TestMockNewWriteRunErrorOutputReadClose(t *testing.T) {
	testNewWriteRunErrorOutputReadClose(t, &mockStore{})
}

func TestFilesNewWriteRunErrorOutputReadClose(t *testing.T) {
	testNewWriteRunErrorOutputReadClose(t, files.NewFiles("."))
}

func TestFilesMultiple(t *testing.T) {
	store := files.NewFiles(".")
	num := 10
	for i := 0; i < (num-1); i++ {
		l, err := NewLog(time.Date(i+10, 0, 0, 0, 0, 0, 0, time.UTC), store)
		if err != nil {
			panic(err)
		}
		l.Write(fmt.Sprintf("%v", i))
		l.Close()
	}
	l, err := NewLog(time.Date(num-1+10, 0, 0, 0, 0, 0, 0, time.UTC), store)
	l.Write(fmt.Sprintf("%v", num-1))
	logs, err := NewLogContents(store)
	if err != nil {
		panic(err)
	}
	if len(logs) != num {
		t.Fatalf("not right number logs expected %v, but have %v", num, len(logs))
	}
	sort.Sort(logs)
	for i, ll := range logs {
		c, err := ll.Open()
		if err != nil {
			panic(err)
		}
		if len(c.Lines) != 1 {
			t.Fatalf("wrong number of lines excpected 1, but have %v", len(c.Lines))
		}
		if c.Lines[0].Line != fmt.Sprintf("%v", num-i-1) {
			t.Fatalf("wrong content expected %v, but have %v", num-i-1, c.Lines[0].Line)
		}
	}
	l.Close()
	for i := 0; i < num; i++ {
		if err := store.DeleteLogSession(time.Date(i+10, 0, 0, 0, 0, 0, 0, time.UTC)); err != nil {
			panic(err)
		}
	}
}


