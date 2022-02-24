package main

import (
	"fmt"
	"github.com/kenlabs/pando/cmd/server/command"
)

func main() {
	rootCmd := command.NewRoot()
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println("Exit with error.")
	}
}
