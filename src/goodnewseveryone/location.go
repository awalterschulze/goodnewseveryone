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
	errUnknownLocation = errors.New("Unknown Location")
)

type LocationId string

type Locations map[LocationId]Location

func configToLocations(log Log, configLoc string) (Locations, error) {
	locations := make(Locations)
	err := filepath.Walk(configLoc, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var loc Location = nil
		if strings.HasSuffix(path, ".remote.json") {
			log.Write(fmt.Sprintf("Remote Config: %v", path))
			loc, err = configToRemoteLocation(path)
			if err != nil {
				return err
			}
			
		} else if strings.HasSuffix(path, ".local.json") {
			log.Write(fmt.Sprintf("Local Config: %v", path))
			loc, err = configToLocalLocation(path)
			if err != nil {
				return err
			}
		}
		if loc == nil {
			return nil
		}
		log.Write(fmt.Sprintf("Location Configured: %v", loc))
		if err := locations.Add(loc); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return locations, nil
}

func (locations Locations) Remove(locId LocationId) error {
	if _, ok := locations[locId]; !ok {
		return errUnknownLocation
	}
	if err := locations[locId].delete(); err != nil {
		return err
	}
	delete(locations, locId)
	return nil
}

func (locations Locations) Add(loc Location) error {
	if _, ok := locations[loc.Id()]; ok {
		return errDuplicateLocation
	}
	err := loc.save()
	if err != nil {
		return err
	}
	locations[loc.Id()] = loc
	return nil
}

func (locations Locations) String() string {
	locs := make([]string, 0, len(locations))
	for _, loc := range locations {
		locs = append(locs, loc.String())
	}
	return "[" + strings.Join(locs, ", ") + "]"
}

type Location interface {
	String() string
	Id() LocationId
	newLocateCommand() *command
	located(log Log, output string) bool
	newMountCommand() *command
	newUmountCommand() *command
	getLocal() string
	save() error
	delete() error
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

func configToRemoteLocation(filename string) (*RemoteLocation, error) {
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

func NewRemoteLocation(typ RemoteLocationType, ipaddress string, mac string, username string, password string, remote string, local string) *RemoteLocation {
	return &RemoteLocation{
		typ,
		ipaddress,
		mac,
		username,
		password,
		remote,
		local,
	}
}

func (this *RemoteLocation) newLocateCommand() *command {
	return newNMapCommand(this.IPAddress)
}

func (this *RemoteLocation) located(log Log, output string) bool {
	if !strings.Contains(output, "Host is up") {
		log.Write(fmt.Sprintf("Cannot Locate %v", this))
		return false
	}
	if !strings.Contains(strings.ToLower(output), strings.ToLower(this.Mac)) {
		log.Write(fmt.Sprintf("Cannot Locate %v", this))
		return false
	}
	return true
}

func (this *RemoteLocation) newMountCommand() *command {
	switch this.Type {
	case FTP:
		return newFTPMountCommand(this.IPAddress, this.Remote, this.Local, this.Username, this.Password)
	case Samba:
		return newCifsMountCommand(this.IPAddress, this.Remote, this.Local, this.Username, this.Password)
	}
	panic("unreachable")
}

func (this *RemoteLocation) newUmountCommand() *command {
	switch this.Type {
	case FTP:
		return newFTPUmountCommand(this.Local)
	case Samba:
		return newCifsUmountCommand(this.Local)
	}
	panic("unreachable")
}

func (this *RemoteLocation) getLocal() string {
	return this.Local
}

func (this *RemoteLocation) String() string {
	return "REMOTE=" + this.Mac + "-" + string(this.Type) + "//" + this.Remote
}

func (this *RemoteLocation) Id() LocationId {
	return LocationId("REMOTE-" + string(this.Type) + "-" +
		strings.Replace(this.Mac, ":", "-", -1) + 
		"-" + 
		strings.Replace(this.Remote, "/", "-", -1))
}

func (this *RemoteLocation) filename() string {
	return fmt.Sprintf("%v.remote.json", this.Id())
}

func (this *RemoteLocation) save() error {
	data, err := json.Marshal(this)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(this.filename(), data, 0666); err != nil {
		return err
	}
	return nil
}

func (this *RemoteLocation) delete() error {
	return os.Remove(this.filename())
}

type LocalLocation struct {
	Local string
}

func configToLocalLocation(filename string) (*LocalLocation, error) {
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

func NewLocalLocation(local string) (*LocalLocation) {
	return &LocalLocation{local}
}

func (this *LocalLocation) String() string {
	return "LOCAL=" + this.Local
}

func (this *LocalLocation) newLocateCommand() *command {
	return nil
}

func (this *LocalLocation) located(log Log, output string) bool {
	return true
}

func (this *LocalLocation) newMountCommand() *command {
	return nil
}

func (this *LocalLocation) newUmountCommand() *command {
	return nil
}

func (this *LocalLocation) getLocal() string {
	return this.Local
}

func (this *LocalLocation) Id() LocationId {
	return LocationId("LOCAL"+strings.Replace(this.Local, "/", "-", -1))
}

func (this *LocalLocation) filename() string {
	return fmt.Sprintf("%v.local.json", this.Id())
}

func (this *LocalLocation) save() error {
	data, err := json.Marshal(this)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(this.filename(), data, 0666); err != nil {
		return err
	}
	return nil
}

func (this *LocalLocation) delete() error {
	return os.Remove(this.filename())
}
