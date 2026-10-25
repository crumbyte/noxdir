package main

import (
	"github.com/crumbyte/noxdir/cmd"
)

var version = "dev"

func main() {
	cmd.Execute(version)
}
