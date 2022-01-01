package main

import (
	"fmt"
	"pando/cmd/client/command"
)

func main() {
	rootCmd := command.NewRoot()
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println("Exit with error")
	}
}
