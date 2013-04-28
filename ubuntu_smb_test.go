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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func testSamba() error {
	conf := `
[global]
   server string = %h server (Samba, Ubuntu)
   usershare allow guests = yes

[testremote]
   comment = my test share
   path = {{.}}
   browsable = yes
   guest ok = yes
   read only = no
   create mask = 0755
`
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	testRemote := filepath.Join(wd, "./testremote/")
	fmt.Printf("remote = %v\n", testRemote)
	testLocal := filepath.Join(wd, "./testlocal/")
	fmt.Printf("local = %v\n", testLocal)
	conf = executeTemplate(conf, testRemote)
	fmt.Printf("Config = %v\n", conf)
	if err := os.Mkdir(testLocal, 0777); err != nil {
		return err
	}
	defer os.RemoveAll(testLocal)
	if err := os.Mkdir(testRemote, 0777); err != nil {
		return err
	}
	defer os.RemoveAll(testRemote)
	thefile := filepath.Join(testRemote, "./professor.txt")
	if err := ioutil.WriteFile(thefile, []byte{1, 2, 3, 4}, 0777); err != nil {
		return err
	}
	defer os.Remove(thefile)
	conffilename := "tiny.conf"
	if err := ioutil.WriteFile(conffilename, []byte(conf), 0777); err != nil {
		return err
	}
	defer os.Remove(conffilename)
	server := exec.Command("smbd", []string{"-iS", "--port=6553", "--configfile=tiny.conf"}...)
	go func() {
		fmt.Printf("server STARTED\n")
		out, err := server.CombinedOutput()
		if err != nil {
			fmt.Printf("ERROR %v\n", err)
		}
		fmt.Printf("server ENDED = %v\n", string(out))
	}()
	example, err := exec.Command("./goodnewseveryone", "-example=true").CombinedOutput()
	fmt.Printf("%v\n", string(example))
	if err != nil {
		return err
	}
	gne := exec.Command("./goodnewseveryone")
	gneIn, err := gne.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		_, err := gneIn.Write(example)
		if err != nil {
			fmt.Printf("Write ERROR: %v\n", err)
		}
		if err := gneIn.Close(); err != nil {
			fmt.Printf("Close ERROR: %v\n", err)
		}
	}()
	output, err := gne.CombinedOutput()
	fmt.Printf("output = %v\n", string(output))
	if err != nil {
		return err
	}
	if !strings.Contains(string(output), "Diff") {
		return errors.New("output does not contain Diff")
	}
	localFiles := createFilelist(testLocal)
	if len(localFiles) != 2 {
		return errors.New(fmt.Sprintf("local should contain one file, but contains %v", localFiles))
	}
	if strings.Contains(localFiles[1], thefile) {
		return errors.New(fmt.Sprintf("want %v got %v", thefile, localFiles[1]))
	}
	return nil
}

func TestSamba(t *testing.T) {
	err := testSamba()
	if err != nil {
		t.Fatalf("%v", err)
	}
}
