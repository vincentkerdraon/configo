package errors

import (
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
	SubCommand subcommand.SubCommand
	Err        error
}

func (err ConfigError) Error() string {
	var res string
	if err.SubCommand != "" {
		res = fmt.Sprintf("on SubCommand: %q,", err.SubCommand)
	}
	return fmt.Sprintf("%sConfigError: %s", res, err.Err)
}

type ParamConfigError struct {
	SubCommand subcommand.SubCommand
	ParamName  paramname.ParamName
	Err        error
}

func (err ParamConfigError) Error() string {
	var res string
	if err.SubCommand != "" {
		res = fmt.Sprintf("on SubCommand: %q,", err.SubCommand)
	}
	return fmt.Sprintf("%sConfigError for Param:%q: %s", res, err.ParamName, err.Err)
}

type ConfigLoadFetchError struct {
	Err error
}

func (err ConfigLoadFetchError) Error() string {
	return fmt.Sprintf("ConfigSyncFetchError: %s", err.Err)
}

type ConfigLoadParseError struct {
	Err error
}

func (err ConfigLoadParseError) Error() string {
	return fmt.Sprintf("ConfigSyncParseError: %s", err.Err)
}

type ConfigWithUsageError struct {
	Err   error
	Usage string
}

func (err ConfigWithUsageError) Error() string {
	return fmt.Sprintf("ConfigWithUsageError: %s\nUsage:\n%s", err.Err, err.Usage)
}
