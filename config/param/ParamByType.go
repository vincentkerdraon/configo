package param

import (
	"strconv"
	"time"

	"github.com/vincentkerdraon/configo/config/param/paramname"
)

//A bunch of helper functions to make life easier.

func NewBool(
	name paramname.ParamName,
	parse func(bool) error,
	opts ...paramOption,
) (*Param, error) {
	return New(name, func(s string) error {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		return parse(b)
	}, opts...)
}

func NewInt(
	name paramname.ParamName,
	parse func(int) error,
	opts ...paramOption,
) (*Param, error) {
	return New(name, func(s string) error {
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		return parse(i)
	}, opts...)
}

func NewInt64(
	name paramname.ParamName,
	parse func(int64) error,
	opts ...paramOption,
) (*Param, error) {
	return New(name, func(s string) error {
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		return parse(int64(i))
	}, opts...)
}

func NewUint(
	name paramname.ParamName,
	parse func(uint) error,
	opts ...paramOption,
) (*Param, error) {
	return New(name, func(s string) error {
		i, err := strconv.ParseUint(s, 10, 0)
		if err != nil {
			return err
		}
		return parse(uint(i))
	}, opts...)
}

func NewUint64(
	name paramname.ParamName,
	parse func(uint64) error,
	opts ...paramOption,
) (*Param, error) {
	return New(name, func(s string) error {
		i, err := strconv.ParseUint(s, 10, 0)
		if err != nil {
			return err
		}
		return parse(i)
	}, opts...)
}

func NewFloat64(
	name paramname.ParamName,
	parse func(float64) error,
	opts ...paramOption,
) (*Param, error) {
	return New(name, func(s string) error {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		return parse(f)
	}, opts...)
}

func NewDuration(
	name paramname.ParamName,
	parse func(time.Duration) error,
	opts ...paramOption,
) (*Param, error) {
	return New(name, func(s string) error {
		d, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		return parse(d)
	}, opts...)
}

func NewString(
	name paramname.ParamName,
	parse func(s string) error,
	opts ...paramOption,
) (*Param, error) {
	return New(name, parse, opts...)
}
