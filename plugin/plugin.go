package plugin

import (
	"github.com/dodo-cli/dodo-core/pkg/plugin"
	"github.com/dodo-cli/dodo-stage-virtualbox/pkg/stage"
)

func RunMe() int {
	plugin.ServePlugins(&stage.Stage{})
	return 0
}

func IncludeMe() {
	plugin.IncludePlugins(&stage.Stage{})
}
