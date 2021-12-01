package config

import (
	"bytes"
	"github.com/mitchellh/go-homedir"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	Convey("test init config", t, func() {
		disSpeed := DisableTestSpeed()
		cfg, err := Init(io.Discard, disSpeed)
		So(err, ShouldBeNil)
		So(cfg.Addresses.P2PAddr, ShouldEqual, defaultP2PAddr)
		So(cfg.Addresses.GraphQL, ShouldEqual, defaultGraphQl)
		So(cfg.Addresses.MetaData, ShouldEqual, defaultMetaData)
		So(cfg.Addresses.GraphSync, ShouldEqual, defaultGraphSync)
		So(len(cfg.AccountLevel.Threshold), ShouldEqual, len(defaultThreshold))
	})
}

func TestSaveLoad(t *testing.T) {
	Convey("test save and load config file", t, func() {
		tmpDir := t.TempDir()
		cfgFile, err := Filename(tmpDir)
		So(err, ShouldBeNil)
		So(filepath.Dir(cfgFile), ShouldEqual, tmpDir)

		disSpeed := DisableTestSpeed()
		cfg, err := Init(io.Discard, disSpeed)
		So(err, ShouldBeNil)

		cfgBytes, err := Marshal(cfg)
		So(err, ShouldBeNil)

		err = cfg.Save(cfgFile)
		So(err, ShouldBeNil)

		cfg2, err := Load(cfgFile)
		So(err, ShouldBeNil)

		cfg2Bytes, err := Marshal(cfg2)
		So(err, ShouldBeNil)

		So(bytes.Equal(cfgBytes, cfg2Bytes), ShouldBeTrue)
	})
	Convey("test load failed", t, func() {
		_, err := Load("just for fun")
		So(err, ShouldEqual, ErrNotInitialized)
	})
}

func TestPath(t *testing.T) {
	Convey("test path", t, func() {
		defDirPath, err := homedir.Expand(DefaultPathRoot)
		So(err, ShouldBeNil)
		p, err := Path("", "1")
		So(err, ShouldBeNil)
		So(p, ShouldEqual, filepath.Join(defDirPath, "1"))

		err = os.Setenv(EnvDir, "???")
		So(err, ShouldBeNil)
		p, err = Path("", "1")
		So(err, ShouldBeNil)
		So(p, ShouldEqual, "???/1")

	})
}
