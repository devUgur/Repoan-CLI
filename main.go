package main

import "github.com/repoan/repoan/cmd"

var version = "0.1.0-dev"

func main() {
	cmd.Execute(version)
}
