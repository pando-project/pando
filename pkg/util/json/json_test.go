package json

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPrettyJsonStr(t *testing.T) {
	Convey("TestPrettyJsonStr", t, func() {
		testStruct := struct {
			Name   string
			Age    int8
			Gender string
		}{
			Name:   "Ben",
			Age:    30,
			Gender: "Male",
		}

		expectStr := `{
    "Name": "Ben",
    "Age": 30,
    "Gender": "Male"
}`

		prettyStr, err := PrettyJsonStr(testStruct)
		if err != nil {
			t.Errorf(err.Error())
		}

		ShouldEqual(prettyStr, expectStr)
	})
}
