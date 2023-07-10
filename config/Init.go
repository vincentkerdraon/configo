package config

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/vincentkerdraon/configo/config/errors"
	"github.com/vincentkerdraon/configo/config/param/paramname"
	"github.com/vincentkerdraon/configo/config/subcommand"
	"golang.org/x/exp/slog"
)

const subCommandLevel0 subcommand.SubCommand = ""

// Init reads the params for the first time and parses the flags
func (c *Manager) Init(ctx context.Context, opts ...configInitOptions) error {
	ci := Init{InputArgs: os.Args[1:]}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&ci); err != nil {
			return err
		}
	}

	//Check and run subCommands. With level0=SubCommand(subCommandLevel0)
	//In some cases, we want to just get the args and ignore completely the commands
	subCommands, args := c.findSubCommand(ci.InputArgs, c.IgnoreCommands)
	c.Logger.DebugCtx(ctx, "findSubCommand and flags", slog.Any("subCommands", subCommands), slog.Any("args", args))
	paramsImpl, initFlags, finalValues, cb, err := c.initParams(ctx, []subcommand.SubCommand{subCommandLevel0}, subCommands, c)
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
		c.Logger.WarnCtx(ctx, "fail parse flags", slog.String("err", err.Error()), slog.Bool("IgnoreFlagProvidedNotDefined", c.IgnoreFlagProvidedNotDefined))
		if !(c.IgnoreFlagProvidedNotDefined && strings.HasPrefix(err.Error(), errFlagProvidedNotDefined)) {
			return c.usageWhenConfigError(errors.ConfigError{SubCommands: subCommands, Err: errors.FlagUnknownError{Err: err}})
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
			c.Logger.DebugCtx(ctx, "Loader will not be synchronizing", slog.String("param", p.Name.String()))
			continue
		}
		if err := c.startSync(ctx, p, c.LoadErrorHandler, append([]subcommand.SubCommand{subCommandLevel0}, subCommands...)); err != nil {
			aggErr.Errs = append(aggErr.Errs, err)
		}
	}

	if aggErr.Errs != nil {
		return c.usageWhenConfigError(aggErr)
	}

	if cb != nil {
		if err := cb(); err != nil {
			return c.usageWhenConfigError(fmt.Errorf("fail command callback, %s", err))
		}
	}
	return nil
}

func (c *Manager) initParams(
	ctx context.Context,
	subCommandsParent []subcommand.SubCommand,
	subCommandsRemaining []subcommand.SubCommand,
	subCmdConfig *Manager,
) (
	_ map[paramname.ParamName]*paramImpl,
	initFlags []func(*flag.FlagSet),
	finalValues []func() (_ error),
	callback func() error,
	_ error,
) {
	paramsImpl := map[paramname.ParamName]*paramImpl{}
	for _, p := range subCmdConfig.Params {
		if p.IsSubCommandLocal && len(subCommandsRemaining) > 0 {
			continue
		}
		pi := &paramImpl{Param: p, hasEnvVarOrFlag: true}
		paramsImpl[p.Name] = pi
		initFlag, setValue, err := pi.init(ctx, c.Logger, c.lock, subCommandsParent)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		initFlags = append(initFlags, initFlag)
		finalValues = append(finalValues, setValue)
	}
	if len(subCommandsRemaining) == 0 {
		return paramsImpl, initFlags, finalValues, subCmdConfig.Callback, nil
	}

	//recursive 1 level down
	subCommandsParent = append(subCommandsParent, subCommandsRemaining[0])
	subSubCmdConfig, ok := subCmdConfig.SubCommands[subCommandsRemaining[0]]
	if !ok {
		expected := []subcommand.SubCommand{}
		for k := range subCmdConfig.SubCommands {
			expected = append(expected, k)
		}
		return nil, nil, nil, nil, errors.ConfigError{
			SubCommands: subCommandsParent,
			Err:         fmt.Errorf("undefined command. Declared: %v", expected)}
	}
	pis, fss, fvs, cb, err := subCmdConfig.initParams(ctx, subCommandsParent, subCommandsRemaining[1:], subSubCmdConfig)
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
	subCommands []subcommand.SubCommand,
) error {
	if p.Loader.Getter == nil {
		return nil
	}
	if p.Loader.SynchroFrequency == 0 {
		return errors.ParamConfigError{ParamName: p.Name, Err: fmt.Errorf("expect SynchroFrequency > 0")}
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
				err := p.load(ctx, c.lock, subCommands)
				if err == nil {
					consecutiveErrNb = 0
					continue
				}
				c.Logger.DebugCtx(ctx, "fail Loader", slog.String("param", p.Name.String()), slog.String("err", err.Error()), slog.Int("consecutiveErrNb", consecutiveErrNb))
				consecutiveErrNb++
				syncError(p.Name, consecutiveErrNb, err)
			}
		}
	}()
	return nil
}

func (c *Manager) findSubCommand(args []string, ignoreCommands bool) (_ []subcommand.SubCommand, argsWithoutCommand []string) {
	res := []subcommand.SubCommand{}
	hasSubCommand := false
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "-") {
			if ignoreCommands {
				continue
			}
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
