package config

import (
	"cuelang.org/go/cue"
	"github.com/hashicorp/go-multierror"
	"github.com/wabenet/dodo-config/pkg/cuetils"
	"github.com/wabenet/dodo-stage/pkg/box"
)

func BoxFromValue(v cue.Value) (*box.Config, error) {
	var errs error

	if out, err := BoxFromStruct(v); err == nil {
		return out, err
	} else {
		errs = multierror.Append(errs, err)
	}

	return nil, errs
}

func BoxFromStruct(v cue.Value) (*box.Config, error) {
	out := &box.Config{}

	if p, ok := cuetils.Get(v, "user"); ok {
		if n, err := StringFromValue(p); err != nil {
			return nil, err
		} else {
			out.User = n
		}
	}

	if p, ok := cuetils.Get(v, "name"); ok {
		if n, err := StringFromValue(p); err != nil {
			return nil, err
		} else {
			out.Name = n
		}
	}

	if p, ok := cuetils.Get(v, "version"); ok {
		if n, err := StringFromValue(p); err != nil {
			return nil, err
		} else {
			out.Version = n
		}
	}

	if p, ok := cuetils.Get(v, "access_token"); ok {
		if n, err := StringFromValue(p); err != nil {
			return nil, err
		} else {
			out.AccessToken = n
		}
	}

	return out, nil
}
