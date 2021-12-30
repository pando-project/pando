package system

import (
	"fmt"
	"os"
)

func Exit(code int, msg string) {
	if code != 0 {
		if msg != "" {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", msg)
		}
		os.Exit(code)
	}

	if msg != "" {
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", msg)
	}
	os.Exit(0)
}
