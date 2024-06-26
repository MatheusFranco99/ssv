package main

import (
	"github.com/MatheusFranco99/ssv/cli"
)

var (
	// AppName is the application name
	AppName = "SSV-Node"

	// Version is the app version
	Version = "latest"
)

func main() {
	cli.Execute(AppName, Version)
}
