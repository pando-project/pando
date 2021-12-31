package client

import "pando/cmd/client/command"

func main() {
	rootCmd := command.NewRoot()
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
