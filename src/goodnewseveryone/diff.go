package goodnewseveryone

import (
	"path/filepath"
	"os"
	"sort"
	"io/ioutil"
	"io"
	"strings"
)

type filelist map[string]bool

func (this filelist) list() []string {
	l := make([]string, 0, len(this))
	for filename, _ := range this {
		l = append(l, filename)
	}
	sort.Strings(l)
	return l
}

func (this filelist) write(writer io.Writer) {
	for _, filename := range this.list() {
		writer.Write([]byte(filename+"\n"))
	}
}

func readFilelist(reader io.Reader) (filelist, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	list := make(filelist)
	for _, line := range lines {
		list[line] = true
	}
	return list, nil
}

func createList(locationKey string) (io.Writer, error) {
	return os.Create(fmt.Sprintf("gne-_-%v-_-%v.list", location, time.Now().Format(DefaultTimeFormat)))
}

type timedFilelist struct {
	locationKey string
	at time.Time
	filename string
}

func newListFiles(root string) {
	filenames := []string{}
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".list") {
			filenames = append(filenames, path)
		}
		return nil
	})
	for _, filename := range filenames {
		f := strings.Split(filename, "-_-")
		if len(f) != 3 {
			continue
		}
		timeStr := strings.Replace(strings.Replace(filename, "gne-", "", 1), ".log", "", 1)
		at, err := time.Parse(strings.Replace(f[2], ".list", "", 1), DefaultTimeFormat)
		if err != nil {
			continue
		}
		&timedFilelist{
			locationKey: f[1],
			at: at,
			filename: filename,
		}
	}
}

func writeList(location Location) (error) {
	list, err := newFilelist(location.getLocal())
	if err != nil {
		return err
	}
	file, err := createList(location.String())
	if err != nil {
		return err
	}
	list.write(file)
	return file.Close()
}

func newFilelist(location string) (filelist, error) {
	files := make(filelist)
	err := filepath.Walk(location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files[path] = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func diffFilelist(oldList filelist, newList filelist) (created filelist, deleted filelist) {
	created = make(filelist)
	deleted = make(filelist)
	for newFile, _ := range newList {
		if _, ok := oldList[newFile]; !ok {
			created[newFile] = true
		}
	}
	for oldFile, _ := range oldList {
		if _, ok := newList[oldFile]; !ok {
			deleted[oldFile] = true
		}
	}
	return
}
