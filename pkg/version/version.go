package version

import "fmt"

var (
	RELEASE = ""

	API = "v1"

	Long = fmt.Sprintf("Pando release: %s, API version: %s", RELEASE, API)
)
