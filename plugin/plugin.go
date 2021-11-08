package plugin

import (
	"github.com/dodo-cli/dodo-core/pkg/plugin"
	"github.com/dodo-cli/dodo-stage-docker-virtualbox/pkg/plugin/stage"
)

func RunMe() int {
	m := plugin.Init()

	m.ServePlugins(stage.New())

	return 0
}

func IncludeMe(m plugin.Manager) {
	m.IncludePlugins(stage.New())
}
