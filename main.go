package main

import (
	"os"

	"github.com/dodo-cli/dodo-stage-virtualbox/plugin"
)

func main() {
	os.Exit(plugin.RunMe())
}
