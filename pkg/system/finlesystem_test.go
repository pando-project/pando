package system

import (
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsDirWritable(t *testing.T) {
	Convey("TestIsDirWritable", t, func() {
		Convey("return false, pathIsEmptyErr if dir is empty", func() {
			writable, err := IsDirWritable("")
			So(writable, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf(pathIsEmptyErr))
		})

		Convey("return false, pathNotExistsErr if path does not exist", func() {
			notExistsPath := "/helloPando"
			writable, err := IsDirWritable("/helloPando")
			So(writable, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf(pathNotExistsErr, notExistsPath))
		})

		Convey("return false, notDirectoryErr if path is not a directory", func() {
			notDirectoryPath := "/bin/sh"
			writable, err := IsDirWritable(notDirectoryPath)
			So(writable, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf(notDirectoryErr, notDirectoryPath))
		})

		Convey("return true, nil if path is really writable", func() {
			writable, err := IsDirWritable("/tmp")
			So(writable, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("return false, nil if path is not writable", func() {
			notWritableDir := "/tmp/not-writable"
			err := os.Mkdir(notWritableDir, 0000)
			if err != nil {
				t.Errorf("try failed to create a test dir(%s): %v", notWritableDir, err)
			}

			writable, err := IsDirWritable(notWritableDir)
			So(writable, ShouldBeFalse)
			So(err, ShouldBeNil)

			err = os.Remove(notWritableDir)
			if err != nil {
				t.Errorf("try failed to cleanup test dir(%s): %v", notWritableDir, err)
			}
		})
	})
}

func TestIsFileExists(t *testing.T) {
	Convey("TestIsFileExists", t, func() {
		Convey("return false, pathIsEmptyErr if file is empty", func() {
			exists, err := IsFileExists("")
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf(pathIsEmptyErr))
		})

		Convey("return false, nil if file does not exist", func() {
			notExistFile := "/hello-pando"
			exists, err := IsFileExists(notExistFile)
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("return false, notFileErr if path is not a file", func() {
			notFilePath := "/tmp"
			exists, err := IsFileExists(notFilePath)
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf(notFileErr, notFilePath))
		})

		Convey("return true, nil if file does exist", func() {
			existFilePath := "/bin/sh"
			exists, err := IsFileExists(existFilePath)
			So(exists, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

func TestIsDirExists(t *testing.T) {
	Convey("TestIsDirExists", t, func() {
		Convey("return false, pathIsEmptyErr if path is empty", func() {
			exists, err := IsDirExists("")
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf(pathIsEmptyErr))
		})

		Convey("return false, nil if dir does not exist", func() {
			exists, err := IsDirExists("/hello-pando")
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("return false, notDirectoryErr if path is not a directory", func() {
			notDirectoryPath := "/bin/sh"
			exists, err := IsDirExists(notDirectoryPath)
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf(notDirectoryErr, notDirectoryPath))
		})

		Convey("return true, nil if dir does exist", func() {
			directoryPath := "/tmp"
			exists, err := IsDirExists(directoryPath)
			So(exists, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}
