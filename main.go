package main

import (
	"os"

	"github.com/wabenet/dodo-stage-docker-virtualbox/pkg/plugin"
)

func main() {
	os.Exit(plugin.RunMe())
}
