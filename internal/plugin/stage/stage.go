package stage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	coreapi "github.com/wabenet/dodo-core/api/core/v1alpha5"
	coreconfig "github.com/wabenet/dodo-core/pkg/config"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-stage-virtualbox/pkg/virtualbox"
	api "github.com/wabenet/dodo-stage/api/stage/v1alpha4"
	"github.com/wabenet/dodo-stage/pkg/plugin/stage"
	"github.com/wabenet/dodo-stage/pkg/util/state"
)

const (
	name           = "virtualbox"
	retryAttempts  = 60
	retrySleeptime = 4
)

var _ stage.Stage = &Stage{}

type Stage struct{}

type State struct {
	IPAddress      string
	Username       string
	PrivateKeyFile string
}

func (s State) MarshalJSON() ([]byte, error) {
	return json.Marshal(s)
}

func New() *Stage {
	return &Stage{}
}

func (vbox *Stage) Type() plugin.Type {
	return stage.Type
}

func (vbox *Stage) PluginInfo() *coreapi.PluginInfo {
	return &coreapi.PluginInfo{
		Name: &coreapi.PluginName{Name: name, Type: stage.Type.String()},
	}
}

func (vbox *Stage) Init() (plugin.Config, error) {
	return map[string]string{}, nil
}

func (vbox *Stage) Cleanup() {}

func (vbox *Stage) Exist(name string) (bool, error) {
	vm := &virtualbox.VM{Name: name}

	if _, err := os.Stat(storagePath(name)); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	_, err := vm.Info()
	return err == nil, nil
}

func (vbox *Stage) Available(name string) (bool, error) {
	vm := &virtualbox.VM{Name: name}

	info, err := vm.Info()
	if err != nil {
		return false, err
	}
	state, ok := info["VMState"]
	return ok && state == "running", nil
}

func (vbox *Stage) GetSSHOptions(name string) (*api.SSHOptions, error) {
	vm := &virtualbox.VM{Name: name}

	portForwardings, err := vm.ListPortForwardings()
	if err != nil {
		return nil, err
	}

	port := 0
	for _, forward := range portForwardings {
		if forward.Name == "ssh" {
			port = forward.HostPort
			break
		}
	}
	if port == 0 {
		return nil, errors.New("no port forwarding matching ssh port found")
	}

	state, err := state.Load[State](name)
	if err != nil {
		return nil, err
	}

	return &api.SSHOptions{
		Hostname:       "127.0.0.1",
		Port:           int32(port),
		Username:       state.Username,
		PrivateKeyFile: state.PrivateKeyFile,
	}, nil
}

func await(test func() (bool, error)) error {
	for attempts := 0; attempts < retryAttempts; attempts++ {
		success, err := test()
		if err != nil {
			return err
		}
		if success {
			return nil
		}
		time.Sleep(retrySleeptime * time.Second)
	}
	return errors.New("max retries reached")
}
func storagePath(name string) string {
	return filepath.Join(coreconfig.GetAppDir(), "stages", name)
}

func persistPath(name string) string {
	return filepath.Join(coreconfig.GetAppDir(), "persist", name)
}
