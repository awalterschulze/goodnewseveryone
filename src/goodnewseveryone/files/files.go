package files

import (
	"goodnewseveryone/store"
	"time"
	"sync"
	"os"
	"sort"
	"path/filepath"
	"io/ioutil"
	"strings"
	"fmt"
)

const defaultTimeFormat = time.RFC3339

const logTimeSep = " | "
const logLineSep = "\n"

type files struct {
	sync.Mutex
	folder string
	openLogFiles map[string]*os.File
	logFiles []string
}

func findLogFiles(root string) []string {
	filenames := []string{}
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".log") {
			filenames = append(filenames, path)
		}
		return nil
	})
	sort.Strings(filenames)
  	return filenames
}

func keyTologFilename(keyStr string) string {
	return fmt.Sprintf("gne-%v.log", keyStr)
}

func logFilenameToKey(filename string) string {
	return strings.Replace(strings.Replace(filename, "gne-", "", 1), ".log", "", 1)
}

func keyToStr(key time.Time) string {
	return key.Format(defaultTimeFormat)
}

func strToKey(keyStr string) (time.Time, error) {
	t, err := time.Parse(defaultTimeFormat, keyStr)
	if err != nil {
		fmt.Printf("time.Parse error = %v\n", err)
	}
	return t, err
}

func NewFiles(folder string) store.Store {
	return &files{
		folder: folder,
		openLogFiles: make(map[string]*os.File),
		logFiles: findLogFiles(folder),
	}
}

func (this *files) NewLogSession(key time.Time) error {
	this.Lock()
	defer this.Unlock()
	keyStr := keyToStr(key)
	if _, ok := this.openLogFiles[keyStr]; ok {
		return store.ErrLogSessionAlreadyExists
	}
	filename := keyTologFilename(keyStr)
	logFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	this.openLogFiles[keyStr] = logFile
	this.logFiles = append(this.logFiles, filename)
	return nil
}

func (this *files) ListLogSessions() []time.Time {
	this.Lock()
	defer this.Unlock()
	times := make([]time.Time, 0, len(this.logFiles))
	for _, filename := range this.logFiles {
		keyStr := logFilenameToKey(filename)
		t, err := strToKey(keyStr)
		if err != nil {
			continue
		}
		times = append(times, t)
	}
	return times
}

func (this *files) ReadFromLogSession(key time.Time) ([]time.Time, []string, error) {
	this.Lock()
	defer this.Unlock()
	keyStr := keyToStr(key)
	filename := keyTologFilename(keyStr)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	datas := strings.Split(string(data), logLineSep)
	times := make([]time.Time, 0, len(datas))
	contents := make([]string, 0, len(datas))
	for _, d := range datas {
		c := strings.Split(d, logTimeSep)
		if len(c) != 2 {
			continue
		}
		t, err := strToKey(c[0])
		if err != nil {
			continue
		}
		times = append(times, t)
		contents = append(contents, c[1])
	}
	return times, contents, nil
}

func (this *files) WriteToLogSession(key time.Time, line string) error {
	this.Lock()
	defer this.Unlock()
	keyStr := keyToStr(key)
	if logFile, ok := this.openLogFiles[keyStr]; ok {
		str := fmt.Sprintf("%v%v%v%v", keyToStr(time.Now()), logTimeSep, line, logLineSep)	
		_, err := logFile.Write([]byte(str))
		if err != nil {
			return err
		}
		return nil
	}
	return store.ErrLogSessionDoesNotExist
}

func (this *files) DeleteLogSession(key time.Time) error {
	this.Lock()
	defer this.Unlock()
	keyStr := keyToStr(key)
	if _, ok := this.openLogFiles[keyStr]; ok {
		return store.ErrLogSessionIsOpenCannotDelete
	}
	filename := keyTologFilename(keyStr)
	if err := os.Remove(filename); err != nil {
		return err
	}
	this.logFiles = findLogFiles(this.folder)
	return nil
}

func (this *files) CloseLogSession(key time.Time) error {
	this.Lock()
	defer this.Unlock()
	keyStr := keyToStr(key)
	if logFile, ok := this.openLogFiles[keyStr]; ok {
		delete(this.openLogFiles, keyStr)
		return logFile.Close()
	}
	return store.ErrLogSessionDoesNotExist
}
