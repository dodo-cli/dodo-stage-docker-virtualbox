package stage

import (
	"os"

	log "github.com/hashicorp/go-hclog"
	"github.com/wabenet/dodo-stage-virtualbox/pkg/virtualbox"
)

func (s *Stage) DeleteStage(name string, force bool, volumes bool) error {
	vm := &virtualbox.VM{Name: name}

	exist, err := s.Exist(name)
	if err != nil {
		if !force {
			return err
		}
	}

	if !exist && !force {
		log.L().Info("VM does not exist")
		return nil
	}

	log.L().Info("removing VM...")

	running, err := s.Available(name)
	if err != nil {
		if !force {
			return err
		}
	}

	if running {
		if err := vm.Stop(true); err != nil {
			if !force {
				return err
			}
		}
	}

	if err = vm.Delete(); err != nil {
		if !force {
			return err
		}
	}

	if err := os.RemoveAll(storagePath(name)); err != nil {
		if !force {
			return err
		}
	}

	if volumes {
		if err := os.RemoveAll(persistPath(name)); err != nil {
			if !force {
				return err
			}
		}
	}

	log.L().Info("removed VM")
	return nil
}
