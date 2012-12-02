package task

import (
	"testing"
	"goodnewseveryone/files"
	"goodnewseveryone/location"
	"reflect"
	"time"
)

func TestTaskTypes(t *testing.T) {
	f := files.NewFiles(".")
	types, err := ListTaskTypes(f)
	if err != nil {
		panic(err)
	}
	if len(types) != 0 {
		t.Fatalf("expected no task types, but read %v", types)
	}
	taskType := TaskType{
		Name: "move",
		CmdStr: "rsync -r --remove-source-files %v %v",
	}
	if err := AddTaskType(f, taskType); err != nil {
		panic(err)
	}
	types, err = ListTaskTypes(f)
	if err != nil {
		panic(err)
	}
	if len(types) != 1 {
		t.Fatalf("expected one type, but read %v", types)
	}
	if !reflect.DeepEqual(types[0], taskType) {
		t.Fatalf("%v != %v", types[0], taskType)
	}
	if err := RemoveTaskType(f, taskType.Name); err != nil {
		panic(err)
	}
	types, err = ListTaskTypes(f)
	if err != nil {
		panic(err)
	}
	if len(types) != 0 {
		t.Fatalf("expected no task types, but read %v", types)
	}
}

func TestTasks(t *testing.T) {
	f := files.NewFiles(".")
	tasks, err := NewTasks(f)
	if err != nil {
		panic(err)
	}
	if len(tasks.tasks) != 0 {
		t.Fatalf("expected no tasks, but read %v", tasks.tasks)
	}
	taskType := TaskType{
		Name: "move",
		CmdStr: "rsync -r --remove-source-files %v %v",
	}
	if err := AddTaskType(f, taskType); err != nil {
		panic(err)
	}
	task := &task{
		typ: taskType,
		src: location.LocationId("Home"),
		dst: location.LocationId("SharedFolder"),
	}
	if err := tasks.Add(task); err != nil {
		panic(err)
	}
	if len(tasks.tasks) != 1 {
		t.Fatalf("expected 1 task, but read %v", tasks.tasks)
	}
	tasks, err = NewTasks(f)
	if err != nil {
		panic(err)
	}
	if len(tasks.tasks) != 1 {
		t.Fatalf("expected 1 task, but read %v", tasks.tasks)
	}
	if err := tasks.Remove(task.Id()); err != nil {
		panic(err)
	}
	if len(tasks.tasks) != 0 {
		t.Fatalf("expected no tasks, but read %v", tasks.tasks)
	}
	tasks, err = NewTasks(f)
	if err != nil {
		panic(err)
	}
	if len(tasks.tasks) != 0 {
		t.Fatalf("expected no tasks, but read %v", tasks.tasks)
	}
	if err := RemoveTaskType(f, taskType.Name); err != nil {
		panic(err)
	}
}

func TestCompleted(t *testing.T) {
	f := files.NewFiles(".")
	taskType := TaskType{
		Name: "move",
		CmdStr: "rsync -r --remove-source-files %v %v",
	}
	if err := AddTaskType(f, taskType); err != nil {
		panic(err)
	}
	tasks, err := NewTasks(f)
	if err != nil {
		panic(err)
	}
	task := &task{
		name: TaskId("moveshared"),
		typ: taskType,
		src: location.LocationId("Home"),
		dst: location.LocationId("SharedFolder"),
	}
	if err := tasks.Add(task); err != nil {
		panic(err)
	}
	if !tasks.tasks[task.Id()].LastCompleted().Equal(time.Time{}) {
		t.Fatalf("%v != %v", time.Time{}, tasks.tasks[task.Id()].LastCompleted())
	}
	now1 := time.Now()
	if err := tasks.Complete(task.Id(), now1); err != nil {
		panic(err)
	}
	now2 := time.Now()
	if err := tasks.Complete(task.Id(), now2); err != nil {
		panic(err)
	}
	if err := tasks.Complete(task.Id(), now1); err != nil {
		panic(err)
	}
	if !tasks.tasks[task.Id()].LastCompleted().Equal(now2) {
		t.Fatalf("%v != %v", now2, tasks.tasks[task.Id()].LastCompleted())
	}
	if err := tasks.Remove(task.Id()); err != nil {
		panic(err)
	}
	if err := RemoveTaskType(f, taskType.Name); err != nil {
		panic(err)
	}
}








