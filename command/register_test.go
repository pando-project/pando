package command

import (
	"Pando/config"
	"context"
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/urfave/cli/v2"
	"os"
	"testing"
)

func TestRegisterWithDifferentInputs(t *testing.T) {
	Convey("when register without peer info then get error", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		tempDir := t.TempDir()
		os.Setenv(config.EnvDir, tempDir)
		app := &cli.App{
			Name: "pando",
			Commands: []*cli.Command{
				RegisterCmd,
			},
		}
		err := app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102"})
		So(err.Error(), ShouldContainSubstring, "please input private key and peerid or config file path")
	})
	Convey("when register with wrong input then get error", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		tempDir := t.TempDir()
		os.Setenv(config.EnvDir, tempDir)

		app := &cli.App{
			Name: "pando",
			Commands: []*cli.Command{
				RegisterCmd,
			},
		}
		wrongpeeridStr := "wrongpeerid"
		peeridStr := "12D3KooWSQJeqeks5YAEzAaLdevYNUXYUp7bk9tHt9UXDQkVS3JC"
		privkeyStr := "CAESQPlmGpltNdirG30V5jH79GrPwp2lJz5rUWC/4leCfYMP9mzDYut2ootG0Xx2ZAy7sGVSsEcVMk0e1+qOOf0KHXU="
		wrongpk := "wrongpk"

		err := app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102", "--privkey", privkeyStr, "--peerid", wrongpeeridStr})
		So(err.Error(), ShouldContainSubstring, "could not decode account id")

		err = app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102", "--privkey", wrongpk, "--peerid", peeridStr})
		So(err.Error(), ShouldContainSubstring, "could not decode private key")

		err = app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102", "--privkey", privkeyStr, "--peerid", "", "-config", "fsdctfvghbjnkctfvgbh"})
		So(err.Error(), ShouldContainSubstring, "no such file or directory")

		wrongIdentity := config.Identity{
			PeerID:  wrongpeeridStr,
			PrivKey: privkeyStr,
		}
		_ = struct {
			Name string
			Age  int
		}{Name: "???", Age: 88}

		f, err := os.Create(tempDir + "testconfig")
		So(err, ShouldBeNil)
		f2, err := os.Create(tempDir + "testconfig2")
		So(err, ShouldBeNil)
		wrongcfgbytes, err := json.Marshal(wrongIdentity)
		So(err, ShouldBeNil)
		_, err = f.Write([]byte("abcdefg"))
		So(err, ShouldBeNil)
		_, err = f2.Write(wrongcfgbytes)
		So(err, ShouldBeNil)

		err = app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102", "--privkey", privkeyStr, "--peerid", "", "-config", tempDir + "testconfig"})
		So(err.Error(), ShouldContainSubstring, "invalid character")

		err = app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102", "--privkey", privkeyStr, "--peerid", "", "-config", tempDir + "testconfig2"})
		So(err.Error(), ShouldContainSubstring, "could not decode account id")

	})
	Convey("when register with proper input and no internet then get error", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		tempDir := t.TempDir()
		os.Setenv(config.EnvDir, tempDir)
		os.Setenv("PANDO", "testurl")

		app := &cli.App{
			Name: "pando",
			Commands: []*cli.Command{
				RegisterCmd,
			},
		}
		peeridStr := "12D3KooWSQJeqeks5YAEzAaLdevYNUXYUp7bk9tHt9UXDQkVS3JC"
		privkeyStr := "CAESQPlmGpltNdirG30V5jH79GrPwp2lJz5rUWC/4leCfYMP9mzDYut2ootG0Xx2ZAy7sGVSsEcVMk0e1+qOOf0KHXU="

		err := app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102", "--privkey", privkeyStr, "--peerid", peeridStr})
		So(err.Error(), ShouldContainSubstring, "failed to register providers")
	})
}
