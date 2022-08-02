package config

import (
	"github.com/hashicorp/go-multierror"
	"github.com/wabenet/dodo-config/pkg/cuetils"
	"github.com/wabenet/dodo-config/pkg/includes"
	"github.com/wabenet/dodo-stage-virtualbox/pkg/spec"
)

func GetAllStages(filenames ...string) (map[string]*Stage, error) {
	var errs error
	stages := map[string]*Stage{}

	resolved, err := includes.ResolveIncludes(filenames...)
	if err != nil {
		errs = multierror.Append(errs, err)
		return stages, errs
	}

	for _, filename := range resolved {
		value, err := cuetils.ReadYAMLFileWithSpec(spec.CueSpec, filename)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		p, ok := cuetils.Get(value, "stages")
		if !ok {
			continue
		}

		s, err := StagesFromValue(p)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		for name, stage := range s {
			stages[name] = stage
		}
	}

	return stages, errs
}
