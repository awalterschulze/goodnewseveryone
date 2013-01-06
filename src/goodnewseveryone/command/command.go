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

package command

import (
	"bufio"
	"bytes"
	"fmt"
	"goodnewseveryone/log"
	"io"
	"os/exec"
	"strings"
	"sync"
)

func read(log log.Log, w io.Writer, r *bufio.Reader) {
	for {
		line, _, err := r.ReadLine()
		if err != nil && err != io.EOF {
			log.Error(err)
			return
		}
		log.Output(line)
		if w != nil {
			w.Write(line)
		}
		if err == io.EOF {
			return
		}
	}
}

type Command interface {
	Run(log log.Log) ([]byte, error)
	Stop(log log.Log)
}

type command struct {
	sync.Mutex
	name         string
	args         []string
	censoredArgs bool
	cmd          *exec.Cmd
}

func NewCommand(f string, args ...string) *command {
	argsinter := make([]interface{}, len(args))
	for i := range args {
		argsinter[i] = args[i]
	}
	c := fmt.Sprintf(f, argsinter...)
	ss := strings.Split(c, " ")
	name := ss[0]
	a := []string{}
	if len(ss) > 1 {
		a = ss[1:]
	}
	return &command{
		name: name,
		args: a,
	}
}

func NewCensoredCommand(f string, args ...string) *command {
	cmd := NewCommand(f, args...)
	cmd.censoredArgs = true
	return cmd
}

func (this *command) Stop(log log.Log) {
	this.Lock()
	defer this.Unlock()
	if this.cmd == nil {
		return
	}
	if this.cmd.Process == nil {
		return
	}
	err := this.cmd.Process.Kill()
	if err != nil {
		log.Error(err)
	}
	return
}

func (this *command) start(log log.Log, output io.Writer) error {
	this.Lock()
	defer this.Unlock()
	if this.censoredArgs {
		log.Run(this.name, "censored arguments")
	} else {
		log.Run(this.name, this.args...)
	}
	this.cmd = exec.Command(this.name, this.args...)
	stdout, err := this.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := this.cmd.StderrPipe()
	if err != nil {
		return err
	}
	outbr := bufio.NewReader(stdout)
	errbr := bufio.NewReader(stderr)
	if err := this.cmd.Start(); err != nil {
		return err
	}
	go read(log, output, outbr)
	go read(log, output, errbr)
	return nil
}

func (this *command) Run(log log.Log) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := this.start(log, buf)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if err := this.cmd.Wait(); err != nil {
		log.Error(err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func NewNMap(ipAddress string) *command {
	return NewCommand("nmap -sP %v", ipAddress)
}

func NewLS(loc string) *command {
	return NewCommand("ls %v", loc)
}

func NewMkdir(loc string) *command {
	return NewCommand("mkdir %v", loc)
}

func NewMount(mount, ipAddress, username, password, remoteLoc, mountLoc string) *command {
	return NewCensoredCommand(mount, username, password, ipAddress, remoteLoc, mountLoc)
}

func NewUnmount(unmount, mountLoc string) *command {
	return NewCommand(unmount, mountLoc)
}
