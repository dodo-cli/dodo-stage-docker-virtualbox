package config

import (
	"cuelang.org/go/cue"
	"github.com/hashicorp/go-multierror"
	"github.com/wabenet/dodo-config/pkg/cuetils"
)

type PersistentVolume struct {
	Size int64
}

func VolumesFromValue(v cue.Value) ([]*PersistentVolume, error) {
	var errs error

	if out, err := VolumesFromMap(v); err == nil {
		return out, err
	} else {
		errs = multierror.Append(errs, err)
	}

	if out, err := VolumesFromList(v); err == nil {
		return out, err
	} else {
		errs = multierror.Append(errs, err)
	}

	return nil, errs
}

func VolumesFromMap(v cue.Value) ([]*PersistentVolume, error) {
	out := []*PersistentVolume{}

	err := cuetils.IterMap(v, func(name string, v cue.Value) error {
		r, err := VolumeFromValue(name, v)
		if err == nil {
			out = append(out, r)
		}

		return err

	})

	return out, err
}

func VolumesFromList(v cue.Value) ([]*PersistentVolume, error) {
	out := []*PersistentVolume{}

	err := cuetils.IterList(v, func(v cue.Value) error {
		r, err := VolumeFromValue("", v)
		if err == nil {
			out = append(out, r)
		}

		return err
	})

	return out, err
}

func VolumeFromValue(name string, v cue.Value) (*PersistentVolume, error) {
	var errs error

	if out, err := VolumeFromStruct(name, v); err == nil {
		return out, err
	} else {
		errs = multierror.Append(errs, err)
	}

	return nil, errs
}

func VolumeFromStruct(name string, v cue.Value) (*PersistentVolume, error) {
	out := &PersistentVolume{}

	if p, ok := cuetils.Get(v, "size"); ok {
		if v, err := BytesFromValue(p); err != nil {
			return nil, err
		} else {
			out.Size = v
		}
	}

	return out, nil
}
