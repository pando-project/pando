package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyJsonStr(t *testing.T) {
	t.Run("TestPrettyJsonStr", func(t *testing.T) {
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

		assert.Equal(t, expectStr, prettyStr)
	})
}
