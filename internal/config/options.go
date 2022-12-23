package config

import (
	"cuelang.org/go/cue"
	"github.com/hashicorp/go-multierror"
	"github.com/wabenet/dodo-config/pkg/cuetils"
)

type Options struct {
	Modify       []string
	Provision    []string
	StagehandURL string
}

func OptionsFromValue(v cue.Value) (*Options, error) {
	var errs error

	if out, err := OptionsFromStruct(v); err == nil {
		return out, err
	} else {
		errs = multierror.Append(errs, err)
	}

	return nil, errs
}

func OptionsFromStruct(v cue.Value) (*Options, error) {
	out := &Options{}

	if p, ok := cuetils.Get(v, "modify"); ok {
		if n, err := StringListFromValue(p); err != nil {
			return nil, err
		} else {
			out.Modify = n
		}
	}

	if p, ok := cuetils.Get(v, "provision"); ok {
		if n, err := StringListFromValue(p); err != nil {
			return nil, err
		} else {
			out.Provision = n
		}
	}

	if p, ok := cuetils.Get(v, "stagehandUrl"); ok {
		if n, err := StringFromValue(p); err != nil {
			return nil, err
		} else {
			out.StagehandURL = n
		}
	}

	return out, nil
}
