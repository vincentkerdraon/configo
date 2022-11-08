package config

import (
	"fmt"
	"os"

	"github.com/vincentkerdraon/configo/config/errors"
	"github.com/vincentkerdraon/configo/config/param"
	"github.com/vincentkerdraon/configo/config/param/paramname"
	"github.com/vincentkerdraon/configo/config/subcommand"
	"github.com/vincentkerdraon/configo/lock"
)

type (
	//Manager contains the parameters definition and does the actual parsing and sync logic.
	Manager struct {
		Params map[paramname.ParamName]param.Param
		// IgnoreFlagProvidedNotDefined when need to ignore some flags.
		//
		// default:false
		IgnoreFlagProvidedNotDefined bool

		Description string

		//LoadErrorHandler is called when an error happens using Loader
		LoadErrorHandler func(_ paramname.ParamName, consecutiveErrNb int, _ error)

		SubCommands map[subcommand.SubCommand]*Manager

		Callback func()

		//lock prevents race condition, mostly when using sync()
		lock lock.Locker
	}

	configOptions func(r *Manager) error
)

// errFlagProvidedNotDefined is a std flag package error. Can only be detected using string prefix
const errFlagProvidedNotDefined = "flag provided but not defined:"

func LoadErrorHandlerDefault(name paramname.ParamName, consecutiveErrNb int, err error) {
	fmt.Printf("fail load param:%q (%d tries), %s", name, consecutiveErrNb, err)
	os.Exit(3)
}

func New(opts ...configOptions) (*Manager, error) {
	c := Manager{
		Params:      make(map[paramname.ParamName]param.Param),
		SubCommands: make(map[subcommand.SubCommand]*Manager),
	}
	for _, opt := range opts {
		if err := opt(&c); err != nil {
			return nil, c.usageWhenConfigError(err)
		}
	}
	if c.LoadErrorHandler == nil {
		c.LoadErrorHandler = LoadErrorHandlerDefault
	}
	if c.lock == nil {
		c.lock = lock.New()
	}
	return &c, nil
}

// WithSubCommand adds a command line subcommands, like `go get` or `git commit`.
//
// Can be called multiple times to define multiple commands, and also recursively to define sub commands.
func WithSubCommand(subCommand subcommand.SubCommand, config *Manager) configOptions {
	return func(c *Manager) error {
		if config == nil {
			return errors.ConfigError{Err: fmt.Errorf("subcommands config can't be nil")}
		}
		if len(subCommand) == 0 {
			return errors.ConfigError{Err: fmt.Errorf("subcommands name can't be empty")}
		}
		if _, f := c.SubCommands[subCommand]; f {
			return errors.ConfigError{Err: fmt.Errorf("2 subcommands have the same name (id)")}
		}
		c.SubCommands[subCommand] = config
		return nil
	}
}

// WithDescription to show in the usage
func WithDescription(d string) configOptions {
	return func(c *Manager) error {
		c.Description = d
		return nil
	}
}

// WithIgnoreFlagProvidedNotDefined when need to ignore some flags.
//
// default:false
func WithIgnoreFlagProvidedNotDefined(t bool) configOptions {
	return func(c *Manager) error {
		c.IgnoreFlagProvidedNotDefined = t
		return nil
	}
}

// WithParamsFromStructTag automatically reads the struct, using struct tags when defined.
//
// Can be called multiple times.
// This uses default options for all params.
// This panics in case of configuration error.
func WithParamsFromStructTag(in interface{}, prefix string) configOptions {
	return func(c *Manager) error {
		params, err := param.ParamsFromStructTag(in, prefix)
		if err != nil {
			panic(err)
		}
		return WithParams(params...)(c)
	}
}

// WithLock for a lock when changing values
func WithLock(l lock.Locker) configOptions {
	return func(c *Manager) error {
		c.lock = l
		return nil
	}
}

// WithParams to set the input parameters
//
// Can be called multiple times.
func WithParams(params ...param.Param) configOptions {
	return func(c *Manager) error {
		for _, p := range params {
			if _, f := c.Params[p.Name]; f {
				return errors.ParamConfigError{ParamName: p.Name, Err: fmt.Errorf("2 params have the same name (id)")}
			}
			c.Params[p.Name] = p
		}
		return nil
	}
}

// WithLoadErrorHandler for error handling. Errors are typed.
//
// Default: Print + os.Exit()
func WithLoadErrorHandler(f func(_ paramname.ParamName, consecutiveErrNb int, _ error)) configOptions {
	return func(c *Manager) error {
		c.LoadErrorHandler = f
		return nil
	}
}

// WithCallback to trigger this function when the parsing is done.
//
// Handy for sub commands.
func WithCallback(f func()) configOptions {
	return func(c *Manager) error {
		c.Callback = f
		return nil
	}
}
