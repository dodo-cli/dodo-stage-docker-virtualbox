package config

import (
	"github.com/dodo-cli/dodo-config/pkg/cuetils"
	"github.com/dodo-cli/dodo-config/pkg/includes"
	"github.com/dodo-cli/dodo-stage-docker-virtualbox/pkg/spec"
	"github.com/hashicorp/go-multierror"
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
