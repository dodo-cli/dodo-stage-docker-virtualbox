package config

import (
	"cuelang.org/go/cue"
	"github.com/hashicorp/go-multierror"
	"github.com/wabenet/dodo-config/pkg/cuetils"
)

type USBFilter struct {
	Name      string
	VendorID  string
	ProductID string
}

func USBFiltersFromValue(v cue.Value) ([]*USBFilter, error) {
	var errs error

	if out, err := USBFiltersFromMap(v); err == nil {
		return out, err
	} else {
		errs = multierror.Append(errs, err)
	}

	if out, err := USBFiltersFromList(v); err == nil {
		return out, err
	} else {
		errs = multierror.Append(errs, err)
	}

	return nil, errs
}

func USBFiltersFromMap(v cue.Value) ([]*USBFilter, error) {
	out := []*USBFilter{}

	err := cuetils.IterMap(v, func(name string, v cue.Value) error {
		r, err := USBFilterFromValue(name, v)
		if err == nil {
			out = append(out, r)
		}

		return err

	})

	return out, err
}

func USBFiltersFromList(v cue.Value) ([]*USBFilter, error) {
	out := []*USBFilter{}

	err := cuetils.IterList(v, func(v cue.Value) error {
		r, err := USBFilterFromValue("", v)
		if err == nil {
			out = append(out, r)
		}

		return err
	})

	return out, err
}

func USBFilterFromValue(name string, v cue.Value) (*USBFilter, error) {
	var errs error

	if out, err := USBFilterFromStruct(name, v); err == nil {
		return out, err
	} else {
		errs = multierror.Append(errs, err)
	}

	return nil, errs
}

func USBFilterFromStruct(name string, v cue.Value) (*USBFilter, error) {
	out := &USBFilter{}

	if p, ok := cuetils.Get(v, "name"); ok {
		if v, err := StringFromValue(p); err != nil {
			return nil, err
		} else {
			out.Name = v
		}
	}

	if p, ok := cuetils.Get(v, "vendorid"); ok {
		if v, err := StringFromValue(p); err != nil {
			return nil, err
		} else {
			out.VendorID = v
		}
	}

	if p, ok := cuetils.Get(v, "productid"); ok {
		if v, err := StringFromValue(p); err != nil {
			return nil, err
		} else {
			out.ProductID = v
		}
	}

	return out, nil
}
