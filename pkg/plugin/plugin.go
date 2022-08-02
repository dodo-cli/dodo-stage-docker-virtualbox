package plugin

import (
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-stage-docker-virtualbox/internal/plugin/stage"
)

func RunMe() int {
	m := plugin.Init()

	m.ServePlugins(stage.New())

	return 0
}

func IncludeMe(m plugin.Manager) {
	m.IncludePlugins(stage.New())
}
