//Copyright 2012 Walter Schulze
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

func createFilelist(location string) []string {
	files := []string{}
	err := filepath.Walk(location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

type filemap map[string]bool

func (this filemap) list(prefix string) []string {
	l := make([]string, 0, len(this))
	for filename, _ := range this {
		l = append(l, prefix+filename)
	}
	sort.Strings(l)
	return l
}

func diffFilelist(oldList []string, newList []string) (created filemap, deleted filemap) {
	created = make(filemap)
	deleted = make(filemap)
	for _, newFile := range newList {
		created[newFile] = true
	}
	for _, oldFile := range oldList {
		if _, ok := created[oldFile]; ok {
			delete(created, oldFile)
		}
	}
	for _, oldFile := range oldList {
		deleted[oldFile] = true
	}
	for _, newFile := range newList {
		if _, ok := deleted[newFile]; ok {
			delete(deleted, newFile)
		}
	}
	return
}

func diff(prevList, curList []string) (created []string, deleted []string) {
	createdList, deletedList := diffFilelist(prevList, curList)
	return createdList.list("+ "), deletedList.list("- ")
}

func run(f string, args ...string) (string, error) {
	if len(f) == 0 {
		return "", nil
	}
	ss := f + " " + strings.Join(args, " ")
	sss := strings.Split(ss, " ")
	name := sss[0]
	a := sss[1:]
	theargs := strings.Split(strings.TrimSpace(strings.Join(a, " ")), " ")
	fmt.Printf("> %v %v\n", name, strings.Join(theargs, " "))
	cmd := exec.Command(name, theargs...)
	output, err := cmd.CombinedOutput()
	fmt.Printf("%v\n", string(output))
	return string(output), err
}

func executeTemplate(glob string, data interface{}) string {
	t := template.Must(template.New("atemplate").Parse(glob))
	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, data); err != nil {
		panic(err)
	}
	return string(buf.Bytes())
}

type Location struct {
	Mount     string
	Unmount   string
	IPAddress string
	Username  string
	Password  string
	Remote    string
	Local     string
}

func (loc *Location) String() string {
	if len(loc.Remote) == 0 {
		return loc.Local
	}
	return fmt.Sprintf("%v/%v", loc.IPAddress, loc.Remote)
}

func (loc *Location) Prepare() {
	if len(loc.Remote) == 0 {
		return
	}
	lsoutput, err := run("ls", loc.Local)
	if err != nil || strings.Contains(string(lsoutput), "No such file or directory") {
		if _, err := run("mkdir", loc.Local); err != nil {
			panic(err)
		}
	}
}

func (loc *Location) GetMount() string {
	if len(loc.Remote) == 0 {
		return ""
	}
	return executeTemplate(loc.Mount, loc)
}

func (loc *Location) GetUnmount() string {
	if len(loc.Remote) == 0 {
		return ""
	}
	return executeTemplate(loc.Unmount, loc)
}

type GoodNews struct {
	Src  *Location
	Dst  *Location
	Task string
}

func (good *GoodNews) Everyone() error {
	task := executeTemplate(good.Task, good)
	run(good.Src.GetUnmount())
	run(good.Dst.GetUnmount())
	good.Src.Prepare()
	good.Dst.Prepare()
	if _, err := run(good.Src.GetMount()); err != nil {
		panic(err)
	}
	defer run(good.Src.GetUnmount())
	if _, err := run(good.Dst.GetMount()); err != nil {
		panic(err)
	}
	defer run(good.Dst.GetUnmount())
	srcPrevList := createFilelist(good.Src.Local)
	dstPrevList := createFilelist(good.Dst.Local)
	defer func() {
		srcNextList := createFilelist(good.Src.Local)
		dstNextList := createFilelist(good.Dst.Local)
		srcCreated, srcDeleted := diff(srcPrevList, srcNextList)
		dstCreated, dstDeleted := diff(dstPrevList, dstNextList)
		fmt.Printf("Diff at %v\n", good.Src.Local)
		fmt.Printf("============================\n")
		fmt.Printf("%v\n", strings.Join(srcCreated, "\n"))
		fmt.Printf("%v\n", strings.Join(srcDeleted, "\n"))
		fmt.Printf("============================\n")
		fmt.Printf("Diff at %v\n", good.Dst.Local)
		fmt.Printf("============================\n")
		fmt.Printf("%v\n", strings.Join(dstCreated, "\n"))
		fmt.Printf("%v\n", strings.Join(dstDeleted, "\n"))
		fmt.Printf("============================\n")
	}()
	_, err := run(task)
	return err
}

func exampleNews() {
	data, err := json.MarshalIndent(&GoodNews{
		Src: &Location{
			Mount:     "mount -o port=6553,guest -t cifs //{{.IPAddress}}/{{.Remote}} {{.Local}}",
			Unmount:   "umount -l {{.Local}}",
			IPAddress: "localhost",
			Username:  "",
			Password:  "",
			Remote:    "testremote",
			Local:     "/media/testremote/",
		},
		Dst: &Location{
			Local: "./testlocal/",
		},
		Task: "rsync -r {{.Src.Local}} {{.Dst.Local}}",
	}, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", string(data))
}

func main() {
	var example bool
	flag.BoolVar(&example, "example", false, "example input folder")
	flag.Parse()
	if example {
		exampleNews()
		return
	}
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	goodnews := &GoodNews{}
	if err := json.Unmarshal(data, goodnews); err != nil {
		panic(err)
	}
	if err := goodnews.Everyone(); err != nil {
		panic(err)
	}
}
