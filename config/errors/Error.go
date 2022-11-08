package errors

import (
	"errors"
	"fmt"

	"github.com/vincentkerdraon/configo/config/param/paramname"
	"github.com/vincentkerdraon/configo/config/subcommand"
)

type ConfigAggregatedError struct {
	Errs []error
}

func (err ConfigAggregatedError) Error() string {
	s := "ConfigAggregatedError:"
	for _, e := range err.Errs {
		s = fmt.Sprintf("%s\n%s", s, e)
	}
	return s
}

type ConfigError struct {
	SubCommands []subcommand.SubCommand
	Err         error
}

func (err ConfigError) Error() string {
	var res string
	if len(err.SubCommands) != 0 {
		if !(len(err.SubCommands) == 1 && err.SubCommands[0] == "") {
			res = fmt.Sprintf("on SubCommands: %v, ", err.SubCommands)
		}
	}
	return fmt.Sprintf("%sConfigError: %s", res, err.Err)
}

type ParamConfigError struct {
	SubCommands []subcommand.SubCommand
	ParamName   paramname.ParamName
	Err         error
}

func (err ParamConfigError) Error() string {
	var res string
	if len(err.SubCommands) != 0 {
		if !(len(err.SubCommands) == 1 && err.SubCommands[0] == "") {
			res = fmt.Sprintf("on SubCommands: %v, ", err.SubCommands)
		}
	}
	return fmt.Sprintf("%sConfigError for Param:%q: %s", res, err.ParamName, err.Err)
}

type ConfigLoaderError struct {
	Err error
}

func (err ConfigLoaderError) Error() string {
	return fmt.Sprintf("ConfigLoaderError: %s", err.Err)
}

type ConfigLoaderFetchError struct {
	Err error
}

func (err ConfigLoaderFetchError) Error() string {
	return fmt.Sprintf("ConfigLoaderFetchError: %s", err.Err)
}

type ConfigWithUsageError struct {
	Err   error
	Usage string
}

func (err ConfigWithUsageError) Error() string {
	return fmt.Sprintf("ConfigWithUsageError: %s\nUsage:\n%s", err.Err, err.Usage)
}

type ParamParseError struct {
	Err error
}

func (err ParamParseError) Error() string {
	return fmt.Sprintf("ParamParseError: %s", err.Err)
}

var MandatoryValueError = errors.New("mandatory value")
var LoaderFetchError = errors.New("fail loader on fetch")
