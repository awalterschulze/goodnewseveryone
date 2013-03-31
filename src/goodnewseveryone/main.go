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
	Type      string
	IPAddress string
	Username  string
	Password  string
	Remote    string
	Local     string
}

func (this *Location) String() string {
	if len(this.Remote) == 0 {
		return this.Local
	}
	return fmt.Sprintf("%v://%v/%v", this.Type, this.IPAddress, this.Remote)
}

func (this *Location) Prepare() {
	if len(this.Remote) == 0 {
		return
	}
	lsoutput, err := run("ls", this.Local)
	if err != nil || strings.Contains(string(lsoutput), "No such file or directory") {
		if _, err := run("mkdir", this.Local); err != nil {
			panic(err)
		}
	}
}

type LocationType struct {
	Name    string
	Mount   string
	Unmount string
}

func (this *LocationType) GetMount(loc *Location) string {
	if len(loc.Remote) == 0 {
		return ""
	}
	return executeTemplate(this.Mount, loc)
}

func (this *LocationType) GetUnmount(loc *Location) string {
	if len(loc.Remote) == 0 {
		return ""
	}
	return executeTemplate(this.Unmount, loc)
}

type LocationTypes []*LocationType

func (this LocationTypes) Get(loc *Location) (mount, unmount string) {
	for i := range this {
		if loc.Type == this[i].Name {
			return this[i].GetMount(loc), this[i].GetUnmount(loc)
		}
	}
	panic(fmt.Sprint("unkown locationType %v", loc.Type))
}

type GoodNews struct {
	News       *News
	Src        string
	SrcMount   string
	SrcUnmount string
	Dst        string
	DstMount   string
	DstUnmount string
}

func MakeItGoodNews(locationTypes LocationTypes, news *News) *GoodNews {
	srcMount, srcUnmount := locationTypes.Get(news.Src)
	dstMount, dstUnmount := locationTypes.Get(news.Dst)
	return &GoodNews{news, news.Src.Local, srcMount, srcUnmount, news.Dst.Local, dstMount, dstUnmount}
}

func (good *GoodNews) Everyone() error {
	task := executeTemplate(good.News.Task, good)
	run(good.SrcUnmount)
	run(good.DstUnmount)
	good.News.Src.Prepare()
	good.News.Dst.Prepare()
	if _, err := run(good.SrcMount); err != nil {
		panic(err)
	}
	defer run(good.SrcUnmount)
	if _, err := run(good.DstMount); err != nil {
		panic(err)
	}
	defer run(good.DstUnmount)
	srcPrevList := createFilelist(good.Src)
	dstPrevList := createFilelist(good.Dst)
	defer func() {
		srcNextList := createFilelist(good.Src)
		dstNextList := createFilelist(good.Dst)
		srcCreated, srcDeleted := diff(srcPrevList, srcNextList)
		dstCreated, dstDeleted := diff(dstPrevList, dstNextList)
		fmt.Printf("Diff at %v\n", good.News.Src)
		fmt.Printf("============================\n")
		fmt.Printf("%v\n", strings.Join(srcCreated, "\n"))
		fmt.Printf("%v\n", strings.Join(srcDeleted, "\n"))
		fmt.Printf("============================\n")
		fmt.Printf("Diff at %v\n", good.News.Dst)
		fmt.Printf("============================\n")
		fmt.Printf("%v\n", strings.Join(dstCreated, "\n"))
		fmt.Printf("%v\n", strings.Join(dstDeleted, "\n"))
		fmt.Printf("============================\n")
	}()
	_, err := run(task)
	return err
}

type News struct {
	Src  *Location
	Dst  *Location
	Task string
}

func exampleNews() {
	data, err := json.MarshalIndent(&News{
		Src: &Location{
			Type:      "smbtest",
			IPAddress: "localhost",
			Username:  "",
			Password:  "",
			Remote:    "testremote",
			Local:     "/media/testremote/",
		},
		Dst: &Location{
			Local: "./testlocal/",
		},
		Task: "rsync -r {{.Src}} {{.Dst}}",
	}, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", string(data))
}

func configLocationTypes(configFolder string) LocationTypes {
	matches, err := filepath.Glob(configFolder + "*.location.json")
	if err != nil {
		panic(err)
	}
	locations := make(LocationTypes, len(matches))
	for i, m := range matches {
		data, err := ioutil.ReadFile(m)
		if err != nil {
			panic(err)
		}
		t := LocationType{}
		if err := json.Unmarshal(data, &t); err != nil {
			panic(err)
		}
		locations[i] = &t
	}
	return locations
}

func main() {
	var example bool
	flag.BoolVar(&example, "example", false, "example input folder")
	var configFolder string
	flag.StringVar(&configFolder, "config", "./", "location of config files")
	flag.Parse()
	if example {
		exampleNews()
		return
	}
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	news := &News{}
	if err := json.Unmarshal(data, news); err != nil {
		panic(err)
	}
	locationTypes := configLocationTypes(configFolder)
	goodnews := MakeItGoodNews(locationTypes, news)
	if err := goodnews.Everyone(); err != nil {
		panic(err)
	}
}
