package system

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

// IsDirWritable checks a directory and return ture if it is writable
func IsDirWritable(dir string) (bool, error) {
	if len(dir) == 0 {
		return false, fmt.Errorf("config directory is not specified, usage: \npando-server init -f config.yaml")
	}

	dirInfo, err := os.Stat(dir)
	if err != nil {
		return false, fmt.Errorf("check directory is writable failed: %v\n", err)
	}

	if !dirInfo.IsDir() {
		return false, fmt.Errorf("%s is not a directory", dir)
	}

	return unix.Access(dir, unix.W_OK) == nil, nil
}

func IsFileExists(file string) (bool, error) {
	if len(file) == 0 {
		return false, fmt.Errorf("file path is not specified.\n")
	}

	fileInfo, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("check file exists failed: %v\n", err)
	}

	if fileInfo.IsDir() {
		return false, fmt.Errorf("the path %s is not a file", file)
	}

	return true, nil
}

func IsDirExists(dir string) (bool, error) {
	if len(dir) == 0 {
		return false, fmt.Errorf("directory path is not specified.\n")
	}

	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("check directory exists failed: %v\n", err)
	}

	return true, nil
}
