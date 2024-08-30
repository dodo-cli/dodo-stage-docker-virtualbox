package stage

import (
	"errors"

	log "github.com/hashicorp/go-hclog"
	"github.com/wabenet/dodo-stage-virtualbox/pkg/virtualbox"
)

func (s *Stage) StopStage(name string) error {
	vm := &virtualbox.VM{Name: name}
	log.L().Info("stopping VM...")

	available, err := s.Available(name)
	if err != nil {
		return err
	}
	if !available {
		log.L().Info("VM is already stopped")
		return nil
	}

	if err := vm.Stop(false); err != nil {
		return err
	}

	if err := await(func() (bool, error) {
		available, err := s.Available(name)
		return !available, err
	}); err != nil {
		return err
	}

	return errors.New("VM did not stop successfully")
}
