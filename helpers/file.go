package helpers

import (
	"os"
	"path/filepath"
)

func OpenOrCreate(dir string) (*os.File, error) {
	file, err := os.OpenFile(dir, os.O_RDWR, 0755)

	if os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(dir), 0755)
		file, err = os.Create(dir)
	}

	if err != nil {
		return nil, err
	}

	return file, nil
}
