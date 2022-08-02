package config

import (
	"cuelang.org/go/cue"
	"github.com/alecthomas/units"
	"github.com/wabenet/dodo-config/pkg/cuetils"
)

func StringFromValue(v cue.Value) (string, error) {
	return v.String()
}

func StringListFromValue(v cue.Value) ([]string, error) {
	out := []string{}

	err := cuetils.IterList(v, func(v cue.Value) error {
		str, err := StringFromValue(v)
		if err == nil {
			out = append(out, str)
		}

		return err
	})

	return out, err
}

func IntFromValue(v cue.Value) (int64, error) {
	return v.Int64()
}

func BytesFromValue(v cue.Value) (int64, error) {
	num, err := v.String()
	if err != nil {
		return 0, err
	}

	return units.ParseStrictBytes(num)
}
