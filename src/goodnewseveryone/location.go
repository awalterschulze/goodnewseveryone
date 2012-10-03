package goodnewseveryone

import (
	"io/ioutil"
	"encoding/json"
	"errors"
	"path/filepath"
	"os"
	"strings"
	"fmt"
)

var (
	errDuplicateLocation = errors.New("Duplicate Location")
)

type locations map[string]Location

func newLocations(log Log, configLoc string) (locations, error) {
	locations := make(locations)
	err := filepath.Walk(configLoc, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".remote.json") {
			log.Write(fmt.Sprintf("Remote Config: %v", path))
			loc, err := ConfigToRemoteLocation(path)
			if err != nil {
				log.Error(err)
				return err
			}
			log.Write(fmt.Sprintf("Location Configured: %v", loc))
			if _, ok := locations[loc.String()]; ok {
				log.Error(errDuplicateLocation)
				return errDuplicateLocation
			}
			locations[loc.String()] = loc
		} else if strings.HasSuffix(path, ".local.json") {
			log.Write(fmt.Sprintf("Local Config: %v", path))
			loc, err := ConfigToLocalLocation(path)
			if err != nil {
				log.Error(err)
				return err
			}
			log.Write(fmt.Sprintf("Location Configured: %v", loc))
			if _, ok := locations[loc.String()]; ok {
				log.Error(errDuplicateLocation)
				return errDuplicateLocation
			}
			locations[loc.String()] = loc
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return locations, nil
}

func (this locations) String() string {
	locs := make([]string, 0, len(this))
	for _, loc := range this {
		locs = append(locs, loc.String())
	}
	return "[" + strings.Join(locs, ", ") + "]"
}

type Location interface {
	String() string
	NewLocateCommand() *command
	Located(log Log, output string) bool
	NewMountCommand() *command
	NewUmountCommand() *command
	GetLocal() string
}

type RemoteLocationType string

var (
	FTP = RemoteLocationType("ftp")
	Samba = RemoteLocationType("smb")
)

var (
	errUndefinedRemoteType = errors.New("Undefined RemoteLocation Type: currently only ftp and smb are supported")
)

type RemoteLocation struct {
	Type RemoteLocationType
	IPAddress string
	Mac string
	Username string
	Password string
	Remote string
	Local string
}

func ConfigToRemoteLocation(filename string) (*RemoteLocation, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	remote := &RemoteLocation{}
	if err := json.Unmarshal(data, &remote); err != nil {
		return nil, err
	}
	if remote.Type != FTP && remote.Type != Samba {
		return nil, errUndefinedRemoteType
	}
	return remote, nil
}

func (this *RemoteLocation) NewLocateCommand() *command {
	return newNMapCommand(this.IPAddress)
}

func (this *RemoteLocation) Located(log Log, output string) bool {
	if !strings.Contains(output, "Host is up") {
		log.Write(fmt.Sprintf("Cannot Locate %v", this))
		return false
	}
	if !strings.Contains(output, this.Mac) {
		log.Write(fmt.Sprintf("Cannot Locate %v", this))
		return false
	}
	return true
}

func (this *RemoteLocation) NewMountCommand() *command {
	switch this.Type {
	case FTP:
		return newFTPMountCommand(this.IPAddress, this.Remote, this.Local, this.Username, this.Password)
	case Samba:
		return newCifsMountCommand(this.IPAddress, this.Remote, this.Local, this.Username, this.Password)
	}
	panic("unreachable")
}

func (this *RemoteLocation) NewUmountCommand() *command {
	switch this.Type {
	case FTP:
		return newFTPUmountCommand(this.Local)
	case Samba:
		return newCifsUmountCommand(this.Local)
	}
	panic("unreachable")
}

func (this *RemoteLocation) GetLocal() string {
	return this.Local
}

func (this *RemoteLocation) String() string {
	return "REMOTE=" + this.Mac + "-" + string(this.Type) + "//" + this.Remote
}

type LocalLocation struct {
	Local string
}

func ConfigToLocalLocation(filename string) (*LocalLocation, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	local := &LocalLocation{}
	if err := json.Unmarshal(data, &local); err != nil {
		return nil, err
	}
	return local, nil
}

func (this *LocalLocation) String() string {
	return "LOCAL=" + this.Local
}

func (this *LocalLocation) NewLocateCommand() *command {
	return nil
}

func (this *LocalLocation) Located(log Log, output string) bool {
	return true
}

func (this *LocalLocation) NewMountCommand() *command {
	return nil
}

func (this *LocalLocation) NewUmountCommand() *command {
	return nil
}

func (this *LocalLocation) GetLocal() string {
	return this.Local
}

