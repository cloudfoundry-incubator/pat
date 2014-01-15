package history

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"time"
)

func LoadAll(baseDir string, as reflect.Type) (all []interface{}, err error) {
	if files, err := loadFiles(baseDir); err == nil {
		return decodeFiles(baseDir, as, files)
	} else {
		return nil, err
	}
}

func LoadBetween(baseDir string, as reflect.Type, from time.Time, to time.Time) (all []interface{}, err error) {
	if files, err := loadFilesBetween(baseDir, from, to); err == nil {
		return decodeFiles(baseDir, as, files)
	} else {
		return nil, err
	}
}

func Save(baseDir string, reply interface{}, timestamp int64) (chain interface{}, err error) {
	encoded, err := json.Marshal(reply)
	if err != nil {
		return nil, err
	}

	if err = createIfNeeded(baseDir); err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path.Join(baseDir, filenameFromTimestamp(timestamp)), encoded, 0644)
	return reply, err
}

func decodeFiles(baseDir string, as reflect.Type, files []os.FileInfo) (all []interface{}, err error) {
	for _, f := range files {
		loaded, err := ioutil.ReadFile(path.Join(baseDir, f.Name()))
		var decoded = reflect.New(as).Interface()
		err = json.Unmarshal(loaded, decoded)
		if err != nil {
			return nil, err
		}

		all = append(all, decoded)
	}

	return all, nil
}

func loadFiles(baseDir string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}
	return files, err
}

func loadFilesBetween(baseDir string, from time.Time, to time.Time) (filtered []os.FileInfo, err error) {
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if i, err := strconv.Atoi(file.Name()); err == nil {
			if from.UnixNano() <= int64(i) && to.UnixNano() >= int64(i) {
				filtered = append(filtered, file)
			}
		} else {
			return nil, err
		}
	}

	return filtered, err
}

func filenameFromTimestamp(timestamp int64) string {
	return fmt.Sprintf("%d", timestamp)
}

func createIfNeeded(dir string) error {
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		fmt.Println("Creating directory, ", dir)
		return os.Mkdir(dir, 0755)
	}
	return nil
}
