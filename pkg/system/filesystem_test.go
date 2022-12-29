package system

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDirWritable(t *testing.T) {
	t.Run("TestIsDirWritable", func(t *testing.T) {
		asserts := assert.New(t)
		t.Run("return false, pathIsEmptyErr if dir is empty", func(t *testing.T) {
			writable, err := IsDirWritable("")
			asserts.False(writable)
			asserts.Equal(fmt.Errorf(pathIsEmptyErr), err)
		})

		t.Run("return false, pathNotExistsErr if path does not exist", func(t *testing.T) {
			notExistsPath := "/helloPando"
			writable, err := IsDirWritable("/helloPando")
			asserts.False(writable)
			asserts.Equal(fmt.Errorf(pathNotExistsErr, notExistsPath), err)
		})

		t.Run("return false, notDirectoryErr if path is not a directory", func(t *testing.T) {
			notDirectoryPath := "/bin/sh"
			writable, err := IsDirWritable(notDirectoryPath)
			asserts.False(writable)
			asserts.Equal(fmt.Errorf(notDirectoryErr, notDirectoryPath), err)
		})

		t.Run("return true, nil if path is really writable", func(t *testing.T) {
			writable, err := IsDirWritable("/tmp")
			asserts.True(writable)
			asserts.Nil(err)
		})

		t.Run("return false, nil if path is not writable", func(t *testing.T) {
			notWritableDir := "/tmp/not-writable"
			err := os.Mkdir(notWritableDir, 0000)
			if err != nil {
				t.Errorf("try failed to create a test dir(%s): %v", notWritableDir, err)
			}

			writable, err := IsDirWritable(notWritableDir)
			asserts.False(writable)
			asserts.Nil(err)

			err = os.Remove(notWritableDir)
			if err != nil {
				t.Errorf("try failed to cleanup test dir(%s): %v", notWritableDir, err)
			}
		})
	})
}

func TestIsFileExists(t *testing.T) {
	t.Run("TestIsFileExists", func(t *testing.T) {
		asserts := assert.New(t)
		t.Run("return false, pathIsEmptyErr if file is empty", func(t *testing.T) {
			exists, err := IsFileExists("")
			asserts.False(exists)
			asserts.Equal(fmt.Errorf(pathIsEmptyErr), err)
		})

		t.Run("return false, nil if file does not exist", func(t *testing.T) {
			notExistFile := "/hello-pando"
			exists, err := IsFileExists(notExistFile)
			asserts.False(exists)
			asserts.Nil(err)
		})

		t.Run("return false, notFileErr if path is not a file", func(t *testing.T) {
			notFilePath := "/tmp"
			exists, err := IsFileExists(notFilePath)
			asserts.False(exists)
			asserts.Equal(fmt.Errorf(notFileErr, notFilePath), err)
		})

		t.Run("return true, nil if file does exist", func(t *testing.T) {
			existFilePath := "/bin/sh"
			exists, err := IsFileExists(existFilePath)
			asserts.True(exists)
			asserts.Nil(err)
		})
	})
}

func TestIsDirExists(t *testing.T) {
	t.Run("TestIsDirExists", func(t *testing.T) {
		asserts := assert.New(t)
		t.Run("return false, pathIsEmptyErr if path is empty", func(t *testing.T) {
			exists, err := IsDirExists("")
			asserts.False(exists)
			asserts.Equal(fmt.Errorf(pathIsEmptyErr), err)
		})

		t.Run("return false, nil if dir does not exist", func(t *testing.T) {
			exists, err := IsDirExists("/hello-pando")
			asserts.False(exists)
			asserts.Nil(err)
		})

		t.Run("return false, notDirectoryErr if path is not a directory", func(t *testing.T) {
			notDirectoryPath := "/bin/sh"
			exists, err := IsDirExists(notDirectoryPath)
			asserts.False(exists)
			asserts.Equal(fmt.Errorf(notDirectoryErr, notDirectoryPath), err)
		})

		t.Run("return true, nil if dir does exist", func(t *testing.T) {
			directoryPath := "/tmp"
			exists, err := IsDirExists(directoryPath)
			asserts.True(exists)
			asserts.Nil(err)
		})
	})
}
