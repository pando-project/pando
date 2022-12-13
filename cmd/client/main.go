package main

import (
	"fmt"
	"github.com/pando-project/pando/cmd/client/command"
)

func main() {
	rootCmd := command.NewRoot()
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println("Exit with error")
	}
}
