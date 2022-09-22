package param

import (
	"fmt"

	"github.com/vincentkerdraon/configo/config/errors"
	"github.com/vincentkerdraon/configo/config/param/paramname"
)

type (
	//Param represents an input parameter.
	Param struct {
		Name paramname.ParamName
		//Parse is the user defined function for this param.
		//Use to decode and set value to a value.
		//Same signature as "Set() error" in std flag package.
		Parse             func(s string) error
		Flag              Flag
		EnvVar            EnvVar
		Loader            Loader
		IsMandatory       bool
		Desc              string
		Examples          []string
		EnumValues        []string
		Default           string
		Exclusive         []paramname.ParamName
		IsSubCommandLocal bool
	}

	paramOption func(*Param) error
)

// New creates a new Param.
func New(
	name paramname.ParamName,
	parse func(s string) error,
	opts ...paramOption,
) (*Param, error) {
	if name == "" {
		return nil, errors.ConfigError{Err: fmt.Errorf("mandatory param name")}
	}
	if parse == nil {
		return nil, errors.ParamConfigError{ParamName: name, Err: fmt.Errorf("mandatory parse function")}
	}
	p := &Param{
		Name:  name,
		Parse: parse,
		EnvVar: EnvVar{
			Use: true,
		},
		Flag: Flag{
			Use: true,
		},
	}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, errors.ParamConfigError{ParamName: name, Err: err}
		}
	}
	return p, nil
}

// WithIsMandatory makes the param value mandatory.
//
// default:false
func WithIsMandatory(t bool) paramOption {
	return func(p *Param) error {
		p.IsMandatory = t
		return nil
	}
}

// WithIsSubCommandLocal makes the param only from this command and not the subcommands.
//
// default:false
// example: `my_program -p1` `my_program subcommand1 -p1`. p1 local to my_program means it won't be automatically detected when using subcommand1.
func WithIsSubCommandLocal(t bool) paramOption {
	return func(p *Param) error {
		p.IsSubCommandLocal = t
		return nil
	}
}

// WithDesc is the param description for showing usage.
func WithDesc(s string) paramOption {
	return func(p *Param) error {
		p.Desc = s
		return nil
	}
}

// WithExamples is an example of possible values for showing usage.
func WithExamples(s ...string) paramOption {
	return func(p *Param) error {
		p.Examples = s
		return nil
	}
}

// WithDefault will be the value if nothing overrides it.
func WithDefault(s string) paramOption {
	return func(p *Param) error {
		p.Default = s
		return nil
	}
}

// WithExclusive defines params that must not have a value if this param is filled.
// Either this param1 or another param2 but not both.
func WithExclusive(params ...paramname.ParamName) paramOption {
	return func(p *Param) error {
		p.Exclusive = params
		return nil
	}
}

// WithEnumValues defines exactly the values that can be use. Anything else will lead to an error.
func WithEnumValues(s ...string) paramOption {
	return func(p *Param) error {
		p.EnumValues = s
		return nil
	}
}
