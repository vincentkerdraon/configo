package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/vincentkerdraon/configo/config/errors"
	"github.com/vincentkerdraon/configo/config/param"
	"github.com/vincentkerdraon/configo/config/subcommand"
)

func TestConfig_usageWhenConfigError(t *testing.T) {
	pCity, err := param.New(
		"City",
		func(s string) error { return nil },
		param.WithDesc("City where user lives"),
		param.WithIsMandatory(true),
		param.WithDefault("Vancouver"),
		param.WithExamples("Toronto", "Vancouver"),
		//When using command line arguments, uses `-Town=` to set this value
		param.WithFlag(param.WithFlagName("Town")),
		//When using environment variables, reads key `TOWN` to set this value
		param.WithEnvVar(param.WithEnvVarName("TOWN")),
		param.WithEnumValues("Toronto", "Vancouver", "Montreal"),
		param.WithExclusive("OtherName"),
		param.WithIsSubCommandLocal(true),
	)
	if err != nil {
		t.Error(err)
	}
	cCity, err := New(WithParams(*pCity), WithDescription("A city reader"))
	if err != nil {
		t.Error(err)
	}

	pAge, err := param.New("Age",
		func(s string) error { return nil },
		param.WithLoader(func(ctx context.Context) (string, error) { return "35", nil }),
		param.WithIsSubCommandLocal(true),
	)
	if err != nil {
		t.Error(err)
	}

	cAgeCommandCity, err := New(WithParams(*pAge), WithSubCommand("City", cCity), WithDescription("An age reader with a command for city"))
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name          string
		config        *Manager
		err           error
		wantErrString string
	}{
		{
			name:   "One param, all options",
			config: cCity,
			err: errors.ParamConfigError{
				SubCommands: []subcommand.SubCommand{subCommandLevel0},
				ParamName:   pCity.Name,
				Err:         fmt.Errorf("err desc"),
			},
			wantErrString: `ConfigWithUsageError: ConfigError for Param:"City": err desc
Usage:
	Param: City
		Description: City where user lives
		Example: [Toronto Vancouver]
		Default: Vancouver
		EnumValues: [Toronto Vancouver Montreal]
		Mandatory value.
		This param won't be available in sub commands.
		Command line flag: -Town
		Environment variable name: Town
		No custom loader defined.
`,
		},
		{
			name:   "Full config",
			config: cCity,
			err: errors.ConfigError{
				SubCommands: []subcommand.SubCommand{subCommandLevel0},
				Err:         fmt.Errorf("err desc"),
			},
			wantErrString: `ConfigWithUsageError: ConfigError: err desc
Usage:

Config/Command description: A city reader

	Param: City
		Description: City where user lives
		Example: [Toronto Vancouver]
		Default: Vancouver
		EnumValues: [Toronto Vancouver Montreal]
		Mandatory value.
		This param won't be available in sub commands.
		Command line flag: -Town
		Environment variable name: Town
		No custom loader defined.

`,
		},
		{
			name:   "Config with sub command",
			config: cAgeCommandCity,
			err: errors.ConfigError{
				SubCommands: []subcommand.SubCommand{subCommandLevel0},
				Err:         fmt.Errorf("err desc"),
			},
			wantErrString: `ConfigWithUsageError: ConfigError: err desc
Usage:

Config/Command description: An age reader with a command for city

	Param: Age
		This param won't be available in sub commands.
		Command line flag: -Age
		Environment variable name: Age
		Using a custom loader without periodic update.

Command: City

	Config/Command description: A city reader

			Param: City
			Description: City where user lives
			Example: [Toronto Vancouver]
			Default: Vancouver
			EnumValues: [Toronto Vancouver Montreal]
			Mandatory value.
			This param won't be available in sub commands.
			Command line flag: -Town
			Environment variable name: Town
			No custom loader defined.


`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.usageWhenConfigError(tt.err); err.Error() != tt.wantErrString {
				t.Errorf("Config.usageWhenConfigError() err string\ngot =%s\ngot =%q\nwant=%q", err.Error(), err.Error(), tt.wantErrString)
			}
		})
	}
}

func TestManagerUsage(t *testing.T) {
	s1p1, err := param.New("p1", func(s string) error { return nil }, param.WithIsMandatory(true))
	if err != nil {
		t.Fatal(err)
	}
	s1, err := New(WithParams(*s1p1))
	if err != nil {
		t.Fatal(err)
	}

	s2p1, err := param.New("p1", func(s string) error { return fmt.Errorf("err parse") })
	if err != nil {
		t.Fatal(err)
	}
	s2, err := New(WithParams(*s2p1))
	if err != nil {
		t.Fatal(err)
	}

	s3p1, err := param.New("p1", func(s string) error { return nil },
		param.WithLoader(func(ctx context.Context) (string, error) { return "", fmt.Errorf("err fetch") }))
	if err != nil {
		t.Fatal(err)
	}
	s3, err := New(WithParams(*s3p1))
	if err != nil {
		t.Fatal(err)
	}

	s4p1, err := param.New("p1", func(s string) error { return fmt.Errorf("err parse") },
		param.WithLoader(func(ctx context.Context) (string, error) { return "val", nil }))
	if err != nil {
		t.Fatal(err)
	}
	s4, err := New(WithParams(*s4p1))
	if err != nil {
		t.Fatal(err)
	}

	s5p1, err := param.New("p1", func(s string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	s5, err := New(WithParams(*s5p1))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		cm          *Manager
		args        []string
		expectedErr error
		check       func()
	}{
		{
			name: "1) missing mandatory param",
			cm:   s1,
			args: []string{},
			expectedErr: errors.ConfigWithUsageError{
				Err: errors.ParamConfigError{
					SubCommands: []subcommand.SubCommand{subCommandLevel0},
					ParamName:   "p1",
					Err:         errors.ErrMandatoryValue,
				},
				Usage: "\tParam: p1\n\t\tMandatory value.\n\t\tCommand line flag: -p1\n\t\tEnvironment variable name: p1\n\t\tNo custom loader defined.\n",
			},
		},
		{
			name: "2) param parse error",
			cm:   s2,
			args: []string{},
			expectedErr: errors.ConfigWithUsageError{
				Err: errors.ParamConfigError{
					SubCommands: []subcommand.SubCommand{subCommandLevel0},
					ParamName:   "p1",
					Err: errors.ParamParseError{
						Err: fmt.Errorf("err parse"),
					},
				},
				Usage: "\tParam: p1\n\t\tCommand line flag: -p1\n\t\tEnvironment variable name: p1\n\t\tNo custom loader defined.\n"},
		},
		{
			name: "3) param load fetch error",
			cm:   s3,
			args: []string{},
			expectedErr: errors.ConfigWithUsageError{
				Err: errors.ParamConfigError{
					SubCommands: []subcommand.SubCommand{subCommandLevel0},
					ParamName:   "p1",
					Err:         errors.ConfigLoaderFetchError{Err: fmt.Errorf("err fetch")},
				},
				Usage: "\tParam: p1\n\t\tCommand line flag: -p1\n\t\tEnvironment variable name: p1\n\t\tUsing a custom loader without periodic update.\n",
			},
		},
		{
			name: "4) param load parse error",
			cm:   s4,
			args: []string{},
			expectedErr: errors.ConfigWithUsageError{
				Err: errors.ParamConfigError{
					SubCommands: []subcommand.SubCommand{subCommandLevel0},
					ParamName:   "p1",
					Err: errors.ParamParseError{
						Err: fmt.Errorf("err parse"),
					},
				},
				Usage: "\tParam: p1\n\t\tCommand line flag: -p1\n\t\tEnvironment variable name: p1\n\t\tUsing a custom loader without periodic update.\n",
			},
		},
		{
			name: "5) unexpected subCommand",
			cm:   s5,
			args: []string{"unexpected"},
			expectedErr: errors.ConfigError{
				SubCommands: []subcommand.SubCommand{subCommandLevel0, "unexpected"},
				Err:         fmt.Errorf("undefined command. Declared: []"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cm.Init(
				context.Background(),
				//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
				WithInputArgs(tt.args),
			)
			if err.Error() != tt.expectedErr.Error() {
				t.Errorf("Config Usage\ngot =%s\ngot =%#v\nwant=%s\nwant=%#v", err, err, tt.expectedErr, tt.expectedErr)
			}
		})
	}
}
