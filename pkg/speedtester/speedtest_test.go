package speedtester

import (
	"testing"
)

func TestFetchInternetSpeed(t *testing.T) {
	downloadSpeed := FetchInternetSpeed(true)
	if downloadSpeed <= 1 {
		t.Errorf("download speed is lower than 1 Mbps, network environment sucks")
	}
}
