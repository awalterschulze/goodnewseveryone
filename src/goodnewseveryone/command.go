package goodnewseveryone

import (
	"os/exec"
	"bufio"
	"sync"
	"io"
	"bytes"
)

func read(log Log, w io.Writer, r *bufio.Reader) {
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

type command struct {
	sync.Mutex
	name string
	args []string
	censoredArgs bool
	cmd *exec.Cmd
}

func newCommand(name string, args ...string) *command {
	return &command{
		name: name,
		args: args,
		censoredArgs: false,
	}
}

func newCensoredCommand(name string, args ...string) *command {
	return &command{
		name: name,
		args: args,
		censoredArgs: true,
	}	
}

func (this *command) stop(log Log) {
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

func (this *command) start(log Log, output io.Writer) error {
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

func (this *command) run(log Log) ([]byte, error) {
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

func newNMapCommand(ipAddress string) *command {
	return newCommand("nmap", "-sP", ipAddress)
}

func newLSCommand(loc string) *command {
	return newCommand("ls", loc)
}

func newMkdirCommand(loc string) *command {
	return newCommand("mkdir", loc)
}

//TODO add mountpoint flag
//http://stephen.rees-carter.net/2011/03/getting-unison-and-samba-to-play-nice/
//http://www.cis.upenn.edu/~bcpierce/unison/download/releases/stable/unison-manual.html#fastcheck
//http://www.cis.upenn.edu/~bcpierce/unison/download/releases/stable/unison-manual.html#mountpoints
//-batch             batch mode: ask no questions at all
func newSyncCommand(loc1, loc2 string) *command {
	return newCommand("unison", "-fastcheck true", "-mountpoint " + loc1, "-mountpoint " + loc2,"-batch", "-dontchmod", "-perms", "0", loc1, loc2)
}

func newBackupCommand(loc1, loc2 string) *command {
	return newCommand("rdiff-backup", loc1, loc2)
}

func newMoveCommand(loc1, loc2 string) *command {
	return newCommand("rsync", "-r", "--remove-source-files", loc1, loc2)
}

func newCifsMountCommand(ipAddress, remoteLoc, mountLoc, username, password string) *command {
	return newCensoredCommand("mount", "-t", "cifs", "//" + ipAddress + "/" + remoteLoc, mountLoc, "-o", "username="+username+",password="+password)
}

func newCifsUmountCommand(loc string) *command {
	return newCommand("umount", "-l", loc)	
}

func newFTPMountCommand(ipAddress, remoteLoc, mountLoc, username, password string) *command {
	return newCensoredCommand("curlftpfs", username+":"+password+"@"+ipAddress+"/"+remoteLoc, mountLoc)
}

func newFTPUmountCommand(loc string) *command {
	return newCommand("fusermount", "-u", loc)
}
