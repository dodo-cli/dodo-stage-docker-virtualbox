package stage

import (
	api "github.com/wabenet/dodo-stage/api/stage/v1alpha4"
)

func (s *Stage) GetStage(name string) (*api.GetStageResponse, error) {
	resp := &api.GetStageResponse{
		Info: &api.StageInfo{
			Name:   name,
			Status: api.StageStatus_NONE,
		},
	}

	exist, err := s.Exist(name)
	if err != nil {
		return nil, err
	}

	if !exist {
		return resp, nil
	}
	resp.Info.Status = api.StageStatus_DOWN

	available, err := s.Available(name)
	if err != nil {
		return nil, err
	}

	if !available {
		return resp, nil
	}
	resp.Info.Status = api.StageStatus_UP

	sshOpts, err := s.GetSSHOptions(name)
	if err != nil {
		return nil, err
	}

	resp.SshOptions = sshOpts

	return resp, nil
}
