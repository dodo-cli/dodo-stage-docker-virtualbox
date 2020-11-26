package stage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

const stateFilename = "state.json"

type State struct {
	IPAddress      string
	Username       string
	PrivateKeyFile string
}

func  loadState(name string) (*State, error) {
	filename := filepath.Join(storagePath(name), stateFilename)

	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		} else {
			return nil, err
		}
	}

	stateFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var state State
	if err := json.Unmarshal(stateFile, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func saveState(name string, state *State) error {
	filename := filepath.Join(storagePath(name), stateFilename)

	stateFile, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, stateFile, 0644)
}
