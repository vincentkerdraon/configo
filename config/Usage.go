package config

import (
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/vincentkerdraon/configo/config/errors"
	"github.com/vincentkerdraon/configo/config/param"
	"github.com/vincentkerdraon/configo/config/param/paramname"
	"github.com/vincentkerdraon/configo/config/subcommand"
)

// Usage displays how to use this configuration.
func (c Manager) Usage(indentation int) string {
	indentString := strings.Repeat("\t", indentation)
	var res string
	append := func(s string) {
		res += "\n" + indentString + s
	}
	if c.Description != "" {
		append("Config/Command description: " + c.Description + "\n")
	} else {
		append("")
	}
	for _, p := range c.Params {
		pi := paramImpl{Param: p}
		append(pi.usage(indentation + 1))
	}
	for command, config := range c.SubCommands {
		append(fmt.Sprintf("Command: %s\n%s", command, config.Usage(indentation+1)))
	}
	return fmt.Sprintf("%s\n", res)
}

// usageWhenConfigError is encapsulating the error to add usage notes.
//
// Best effort. Retuning directly the err if can't link it to a param or config.
func (c *Manager) usageWhenConfigError(err error) error {
	pce := errors.ParamConfigError{}
	if stderrors.As(err, &pce) {
		p := c.getParamInSubCommands(pce.SubCommands, pce.ParamName)
		if p == nil {
			return err
		}
		pi := paramImpl{Param: *p}
		return errors.ConfigWithUsageError{
			Err:   err,
			Usage: pi.usage(1),
		}
	}
	ce := errors.ConfigError{}
	if stderrors.As(err, &ce) {
		cmd := c.getSubCommand(ce.SubCommands)
		if cmd == nil {
			flagUnknownError := errors.FlagUnknownError{}
			if stderrors.As(err, &flagUnknownError) {
				return errors.ConfigWithUsageError{
					Err:   err,
					Usage: c.Usage(0),
				}
			}
			return err
		}
		return errors.ConfigWithUsageError{
			Err:   err,
			Usage: cmd.Usage(0),
		}
	}
	return err
}

func (c *Manager) getParamInSubCommands(subCommands []subcommand.SubCommand, paramName paramname.ParamName) *param.Param {
	var m *Manager = c
	for _, subCmd := range subCommands {
		if subCmd != subCommandLevel0 {
			m = m.SubCommands[subCmd]
		}
		p, f := m.Params[paramName]
		if f {
			return &p
		}
	}
	return nil
}

func (c *Manager) getSubCommand(subCommands []subcommand.SubCommand) *Manager {
	var res *Manager
	for _, subCmd := range subCommands {
		if subCmd == "" {
			res = c
			continue
		}
		res = c.SubCommands[subCmd]
	}
	return res
}
