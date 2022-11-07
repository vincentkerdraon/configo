package config

import (
	"context"
	stderrors "errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/vincentkerdraon/configo/config/errors"
	"github.com/vincentkerdraon/configo/config/param/paramname"
	"github.com/vincentkerdraon/configo/config/subcommand"
)

// Init will read the params for the first time and parse the flags
func (c *Manager) Init(ctx context.Context, opts ...configInitOptions) error {
	ci := Init{InputArgs: os.Args[1:]}
	for _, opt := range opts {
		if err := opt(&ci); err != nil {
			return err
		}
	}

	//Check and run subCommands. With level0=SubCommand("")
	subCommands, args := c.findSubCommand(ci.InputArgs)
	subCommands = append([]subcommand.SubCommand{""}, subCommands...)
	paramsImpl, initFlags, finalValues, cb, err := c.initParams(ctx, subCommands, c)
	if err != nil {
		return c.usageWhenConfigError(err)
	}

	//using the std go flag lib is a bit annoying. But std are good.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	for _, initFlag := range initFlags {
		initFlag(fs)
	}
	if err := fs.Parse(args); err != nil {
		if !(c.IgnoreFlagProvidedNotDefined && strings.HasPrefix(err.Error(), errFlagProvidedNotDefined)) {
			return c.usageWhenConfigError(errors.ConfigError{Err: err})
		}
	}

	//Now set the destination value once.
	for _, fv := range finalValues {
		if err := fv(); err != nil {
			return c.usageWhenConfigError(err)
		}
	}
	var aggErr errors.ConfigAggregatedError

	//Check exclusive params
	for _, p := range paramsImpl {
		for _, excl := range p.Exclusive {
			p2, ok := paramsImpl[excl]
			if !ok {
				//it should exist! check done at creation.
				continue
			}
			if p2.hasValue {
				aggErr.Errs = append(aggErr.Errs, errors.ParamConfigError{ParamName: p.Name, Err: fmt.Errorf("exclusive with param:%q", p2.Name)})
			}
		}
	}

	//Start sync. Skip if not defined or if has EnvVar or Flag override. (Loader is lower priority)
	for _, p := range paramsImpl {
		if p.Loader.Getter == nil || p.Loader.SynchroFrequency == 0 || p.hasEnvVarOrFlag {
			continue
		}
		if err := c.startSync(ctx, p, c.LoadErrorHandler); err != nil {
			aggErr.Errs = append(aggErr.Errs, err)
		}
	}

	if aggErr.Errs != nil {
		return c.usageWhenConfigError(aggErr)
	}

	if cb != nil {
		cb() //FIXME test that
	}
	return nil
}

// usageWhenConfigError is encapsulating the error to add usage notes.
func (c *Manager) usageWhenConfigError(err error) error {
	pce := errors.ParamConfigError{}
	if stderrors.As(err, &pce) {
		p, f := c.Params[pce.ParamName]
		if !f {
			return err
		}
		pi := paramImpl{Param: p}
		return errors.ConfigWithUsageError{
			Err:   err,
			Usage: pi.usage(1),
		}
	}
	ce := errors.ConfigError{}
	if stderrors.As(err, &ce) {
		return errors.ConfigWithUsageError{
			Err:   err,
			Usage: c.Usage(0),
		}
	}
	return err
}

func (c Manager) Usage(indent int) string {
	indentString := strings.Repeat("\t", indent)
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
		append(pi.usage(indent + 1))
	}
	for command, config := range c.SubCommand {
		append(fmt.Sprintf("Command: %s\n%s", command, config.Usage(indent+1)))
	}
	return fmt.Sprintf("%s\n", res)
}

func (c *Manager) initParams(
	ctx context.Context,
	subCmd []subcommand.SubCommand,
	subCmdConfig *Manager,
) (
	_ map[paramname.ParamName]*paramImpl,
	initFlags []func(*flag.FlagSet),
	finalValues []func() (_ error),
	callback func(),
	_ error,
) {
	paramsImpl := map[paramname.ParamName]*paramImpl{}
	for _, p := range subCmdConfig.Params {
		if p.IsSubCommandLocal && len(subCmd) > 1 {
			continue
		}
		pi := &paramImpl{Param: p, hasEnvVarOrFlag: true}
		paramsImpl[p.Name] = pi
		initFlag, setValue, err := pi.init(ctx, c.lock)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		initFlags = append(initFlags, initFlag)
		finalValues = append(finalValues, setValue)
	}
	if len(subCmd) == 1 {
		return paramsImpl, initFlags, finalValues, subCmdConfig.Callback, nil
	}
	//recursive 1 level down
	subSubCmdConfig, ok := subCmdConfig.SubCommand[subCmd[1]]
	if !ok {
		expected := []subcommand.SubCommand{}
		for k := range subCmdConfig.SubCommand {
			expected = append(expected, k)
		}
		return nil, nil, nil, nil, errors.ConfigError{
			SubCommand: subCmd[1],
			Err:        fmt.Errorf("undefined command. Declared %v", expected)}
	}
	pis, fss, fvs, cb, err := subCmdConfig.initParams(ctx, subCmd[1:], subSubCmdConfig)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for k, v := range paramsImpl {
		pis[k] = v
	}

	return pis, append(fss, initFlags...), append(fvs, finalValues...), cb, nil
}

func (c *Manager) startSync(
	ctx context.Context,
	p *paramImpl,
	syncError func(_ paramname.ParamName, consecutiveErrNb int, _ error),
) error {
	if p.Loader.Getter == nil {
		return nil
	}
	if p.Loader.SynchroFrequency == 0 {
		return errors.ParamConfigError{ParamName: p.Name, Err: fmt.Errorf("expect sync freq > 0")}
	}
	go func() {
		ticker := time.NewTicker(p.Loader.SynchroFrequency)
		defer ticker.Stop()
		consecutiveErrNb := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := p.load(ctx, c.lock)
				if err == nil {
					consecutiveErrNb = 0
					continue
				}
				consecutiveErrNb++
				syncError(p.Name, consecutiveErrNb, err)
			}
		}
	}()
	return nil
}

// ForceLoad will immediately sync all the params where Load() is defined
func (c *Manager) ForceLoad(ctx context.Context) error {
	for _, p := range c.Params {
		pi := paramImpl{Param: p}
		if err := pi.load(ctx, c.lock); err != nil {
			return err
		}
	}
	return nil
}

func (c *Manager) findSubCommand(args []string) (_ []subcommand.SubCommand, argsWithoutCommand []string) {
	res := []subcommand.SubCommand{}
	hasSubCommand := false
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "-") {
			hasSubCommand = true
			res = append(res, subcommand.SubCommand(args[i]))
		} else {
			return res, args[i:]
		}
	}
	if hasSubCommand {
		return res, []string{}
	}
	return res, args
}
