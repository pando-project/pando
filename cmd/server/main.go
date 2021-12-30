package main

import "pando/cmd/server/command"

func main() {
	rootCmd := command.NewRoot()
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
