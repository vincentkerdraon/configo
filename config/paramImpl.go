package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/vincentkerdraon/configo/config/errors"
	"github.com/vincentkerdraon/configo/config/param"
	"github.com/vincentkerdraon/configo/config/subcommand"
	"github.com/vincentkerdraon/configo/lock"
	"golang.org/x/exp/slog"
)

//Keeping this implementation non-exported to keep the public API clean

type paramImpl struct {
	param.Param

	// hasValue to check if mandatory values are set.
	//
	// internal
	hasValue bool

	// hasEnvVarOrFlag to start the sync or not (depending of system override or not)
	//
	// internal
	hasEnvVarOrFlag bool
}

func (p *paramImpl) init(ctx context.Context, logger *slog.Logger, lock lock.Locker, subCommands []subcommand.SubCommand) (initFlag func(*flag.FlagSet), setValue func() error, _ error) {
	var hasEnvVarOrFlag bool
	val := p.Default

	var valEnvVar string
	if p.EnvVar.Use {
		valEnvVar = p.loadEnvVar()
		if valEnvVar != "" {
			hasEnvVarOrFlag = true
			val = valEnvVar
			logger.DebugCtx(ctx, "found env var", slog.String("Param", p.Name.String()), slog.String("Value", val))
		} else {
			logger.DebugCtx(ctx, "no env var found", slog.String("Param", p.Name.String()))
		}
	}
	tmpFlagVal := val
	if p.Flag.Use {
		//using tmpFlagVal to detect if the value was set or not (even when same as previous step)
		initFlag = p.loadFlag(logger, &tmpFlagVal)
	}
	setValue = func() error {
		if tmpFlagVal != val {
			hasEnvVarOrFlag = true
			val = tmpFlagVal
		}

		if !hasEnvVarOrFlag {
			if p.Loader.Getter != nil {
				valLoader, err := p.Loader.Getter(ctx)
				if err != nil {
					return errors.ParamConfigError{ParamName: p.Name, SubCommands: subCommands, Err: errors.ConfigLoaderFetchError{Err: err}}
				}
				if valLoader != "" {
					val = valLoader
					logger.DebugCtx(ctx, "Loader returns value", slog.String("Param", p.Name.String()), slog.String("Value", val))
				} else {
					logger.DebugCtx(ctx, "Loader returns no value", slog.String("Param", p.Name.String()))
				}
			}
		} else {
			logger.DebugCtx(ctx, "skipping Loader, found env var or flag", slog.String("Param", p.Name.String()), slog.String("Value", val))
		}

		//Check mandatory
		if p.IsMandatory && val == "" {
			return errors.ParamConfigError{ParamName: p.Name, SubCommands: subCommands, Err: errors.ErrMandatoryValue}
		}

		//check enum
		if err := p.checkEnum(val); err != nil {
			return errors.ParamConfigError{ParamName: p.Name, SubCommands: subCommands, Err: err}
		}

		//Check if flag or envvar set. In this case, don't start sync.
		p.hasEnvVarOrFlag = hasEnvVarOrFlag
		//Check exclusive values
		p.hasValue = (val != "")

		return p.lockAndParse(ctx, lock, val, subCommands)
	}

	return initFlag, setValue, nil
}

func (p paramImpl) checkEnum(val string) error {
	if len(p.EnumValues) == 0 {
		return nil
	}
	for _, v := range p.EnumValues {
		if v == val {
			return nil
		}
	}
	return fmt.Errorf("got value:%q, expect one of:%v", val, p.EnumValues)
}

func (p paramImpl) usage(indent int) string {
	indentString := strings.Repeat("\t", indent)
	res := indentString + "Param: " + p.Name.String()
	append := func(s string) {
		res += "\n" + indentString + "\t" + s
	}
	if len(p.Desc) > 0 {
		append("Description: " + p.Desc)
	}
	if len(p.Examples) > 0 {
		append(fmt.Sprintf("Example: %v", p.Examples))
	}
	if p.Default != "" {
		append("Default: " + p.Default)
	}
	if len(p.EnumValues) > 0 {
		append(fmt.Sprintf("EnumValues: %v", p.EnumValues))
	}
	if p.IsMandatory {
		append("Mandatory value.")
	}
	if p.IsSubCommandLocal {
		append("This param won't be available in sub commands.")
	}
	if p.Flag.Use {
		if p.Flag.Name == "" {
			append("Command line flag: -" + p.Name.String())
		} else {
			append("Command line flag: -" + p.Flag.Name)
		}
	} else {
		append("Command line flag disable.")
	}
	if p.EnvVar.Use {
		if p.EnvVar.Name == "" {
			append("Environment variable name: " + p.Name.String())
		} else {
			append("Environment variable name: " + p.Flag.Name)
		}
	} else {
		append("Environment variable disable.")
	}
	if p.Loader.Getter != nil {
		if p.Loader.SynchroFrequency == 0 {
			append("Using a custom loader without periodic update.")
		} else {
			append("Using a custom loader, refresh every " + p.Loader.SynchroFrequency.String())
		}
	} else {
		append("No custom loader defined.")
	}

	return res + "\n"
}

func (p paramImpl) loadEnvVar() string {
	var nameEnvVar string
	if p.EnvVar.Name != "" {
		nameEnvVar = p.EnvVar.Name
	} else {
		nameEnvVar = p.Name.String()
	}
	return os.Getenv(nameEnvVar)
}

func (p paramImpl) loadFlag(logger *slog.Logger, val *string) func(*flag.FlagSet) {
	var nameFlag string
	if p.Flag.Name != "" {
		nameFlag = p.Flag.Name
	} else {
		nameFlag = p.Name.String()
	}

	return func(fs *flag.FlagSet) {
		logger.Debug("Param:%s, checking flag named %q", p.Name, nameFlag)
		fs.StringVar(val, nameFlag, *val, p.usage(0))
	}
}

func (p paramImpl) load(ctx context.Context, lock lock.Locker, valuePrevious string, subCommands []subcommand.SubCommand) (valueNew string, _ error) {
	if p.Loader.Getter == nil {
		return "", nil
	}

	val, err := p.Loader.Getter(ctx)
	if err != nil {
		return "", errors.ConfigLoaderError{Err: errors.ConfigLoaderFetchError{Err: err}}
	}
	if valuePrevious == val {
		return valuePrevious, nil
	}
	if err := p.lockAndParse(ctx, lock, val, subCommands); err != nil {
		return "", errors.ConfigLoaderError{Err: err}
	}
	return val, nil
}

func (p *paramImpl) lockAndParse(ctx context.Context, lock lock.Locker, s string, subCommands []subcommand.SubCommand) error {
	//Because the value is set using outside code, we don't know if it is always quick.
	//Adding a protection where Timeout can be used.
	if err := lock.LockWithContext(ctx); err != nil {
		return err
	}
	defer lock.Unlock()

	err := p.Parse(s)
	if err != nil {
		return errors.ParamConfigError{ParamName: p.Name, SubCommands: subCommands, Err: errors.ParamParseError{Err: err}}
	}
	return nil
}
