package option

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDiscoveryFormat(t *testing.T) {
	t.Run("test discovery format", func(t *testing.T) {
		durationText := "1s"
		ds := &Discovery{
			PollInterval:   durationText,
			RediscoverWait: durationText,
			Timeout:        durationText,
		}
		d := ds.PollIntervalInDurationFormat()
		asserts := assert.New(t)
		asserts.Equal(Duration(time.Second), d)
		d = ds.RediscoverWaitInDurationFormat()
		asserts.Equal(Duration(time.Second), d)
		d = ds.TimeoutInDurationFormat()
		asserts.Equal(Duration(time.Second), d)
	})
}
