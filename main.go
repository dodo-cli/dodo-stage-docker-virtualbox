package main

import (
	"os"

	"github.com/dodo-cli/dodo-stage-docker-virtualbox/plugin"
)

func main() {
	os.Exit(plugin.RunMe())
}
