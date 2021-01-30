package plugin

import (
	"github.com/dodo-cli/dodo-core/pkg/plugin"
	"github.com/dodo-cli/dodo-stage-docker-virtualbox/pkg/stage"
)

func RunMe() int {
	plugin.ServePlugins(&stage.Stage{})
	return 0
}

func IncludeMe() {
	plugin.IncludePlugins(&stage.Stage{})
}
