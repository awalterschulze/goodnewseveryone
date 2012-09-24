package goodnewseveryone

import (
	"path/filepath"
	"os"
)

func list(location string) (map[string]bool, error) {
	files := make(map[string]bool)
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

func diff(oldList map[string]bool, newList map[string]bool) (created []string, deleted []string) {
	for newFile, _ := range newList {
		if _, ok := oldList[newFile]; !ok {
			created = append(created, newFile)
		}
	}
	for oldFile, _ := range oldList {
		if _, ok := newList[oldFile]; !ok {
			deleted = append(deleted, oldFile)
		}
	}
	return
}
