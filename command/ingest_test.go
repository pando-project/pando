package command

import (
	"context"
	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/urfave/cli/v2"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"
)

var app = &cli.App{
	Name: "indexer",
	Commands: []*cli.Command{
		IngestCmd,
	},
}

func TestMissingParam(t *testing.T) {
	Convey("when input missing parameter then get error", t, func() {
		ctx := context.Background()
		os.Setenv("PANDO", "testurl")
		err := app.RunContext(ctx, []string{"pando", "ingest", "subscribe"})
		So(ShouldContainSubstring(err.Error(), "Required flag \"prov\" not set"), ShouldBeEmpty)

		err = app.RunContext(ctx, []string{"pando", "ingest", "subscribe", "-prov"})
		So(ShouldContainSubstring(err.Error(), "flag needs an argument: -prov"), ShouldBeEmpty)

		err = app.RunContext(ctx, []string{"pando", "ingest", "unsubscribe"})
		So(ShouldContainSubstring(err.Error(), "Required flag \"prov\" not set"), ShouldBeEmpty)

		err = app.RunContext(ctx, []string{"pando", "ingest", "unsubscribe", "-prov"})
		So(ShouldContainSubstring(err.Error(), "flag needs an argument: -prov"), ShouldBeEmpty)

	})

}

func TestWrongParam(t *testing.T) {
	Convey("when input wrong parameter then get error", t, func() {
		ctx := context.Background()
		os.Setenv("PANDO", "testurl")

		err := app.RunContext(ctx, []string{"pando", "ingest", "subscribe", "--prov", "?????/"})
		So(err.Error(), ShouldContainSubstring, "failed to parse peer ID")

		err = app.RunContext(ctx, []string{"pando", "ingest", "unsubscribe", "--prov", "?????/"})
		So(err.Error(), ShouldContainSubstring, "failed to parse peer ID")

	})

}

func TestNormal(t *testing.T) {
	Convey("when input proper parameter then get nil error", t, func() {
		ctx := context.Background()
		os.Setenv("PANDO", "testurl")
		res1 := ApplyMethod(reflect.TypeOf(&http.Client{}), "Do", func(client *http.Client, _ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       http.NoBody,
			}, nil
		})
		ApplyFunc(io.ReadAll, func(_ io.Reader) ([]byte, error) {
			return []byte("test response"), nil
		})
		err := app.RunContext(ctx, []string{"pando", "ingest", "subscribe", "--prov", "12D3KooWSQJeqeks5YAEzAaLdevYNUXYUp7bk9tHt9UXDQkVS3JC"})
		So(err, ShouldBeNil)
		res1.Reset()

		res2 := ApplyMethod(reflect.TypeOf(&http.Client{}), "Do", func(client *http.Client, _ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Body:       http.NoBody,
			}, nil
		})

		err = app.RunContext(ctx, []string{"pando", "ingest", "subscribe", "--prov", "12D3KooWSQJeqeks5YAEzAaLdevYNUXYUp7bk9tHt9UXDQkVS3JC"})
		So(err.Error(), ShouldContainSubstring, "500 Internal Server Error")
		res2.Reset()

	})

}
