package option

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDuration(t *testing.T) {
	t.Run("TestDuration", func(t *testing.T) {
		t.Run("Test Duration helper functions", func(t *testing.T) {
			duration := Duration(0)
			durationText := "1s"
			err := duration.UnmarshalText([]byte(durationText))
			if err != nil {
				t.Error(err)
			}
			durationBytes, err := duration.MarshalText()
			if err != nil {
				t.Error(err)
			}
			asserts := assert.New(t)
			asserts.Equal([]byte(durationText), durationBytes)
			asserts.Equal(durationText, duration.String())
		})
	})
}
