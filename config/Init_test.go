package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/vincentkerdraon/configo/config/errors"
	"github.com/vincentkerdraon/configo/config/param"
	"github.com/vincentkerdraon/configo/config/param/paramname"
	"github.com/vincentkerdraon/configo/config/subcommand"
)

func TestConfig_findSubCommand(t *testing.T) {
	tests := []struct {
		name                   string
		args                   []string
		wantSubCommands        []subcommand.SubCommand
		wantArgsWithoutCommand []string
	}{
		{
			name:                   "ok",
			args:                   []string{"sc1", "sc2", "-p1", "-p2"},
			wantSubCommands:        []subcommand.SubCommand{"sc1", "sc2"},
			wantArgsWithoutCommand: []string{"-p1", "-p2"},
		},
		{
			name:                   "empty",
			args:                   []string{},
			wantSubCommands:        []subcommand.SubCommand{},
			wantArgsWithoutCommand: []string{},
		},
		{
			name:                   "no sub cmd",
			args:                   []string{"-p1", "-p2"},
			wantSubCommands:        []subcommand.SubCommand{},
			wantArgsWithoutCommand: []string{"-p1", "-p2"},
		},
		{
			name:                   "only sub cmd",
			args:                   []string{"sc1", "sc2"},
			wantSubCommands:        []subcommand.SubCommand{"sc1", "sc2"},
			wantArgsWithoutCommand: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Manager{}
			gotSubCommands, gotArgsWithoutCommand := c.findSubCommand(tt.args)
			if fmt.Sprintf("%+v", gotSubCommands) != fmt.Sprintf("%+v", tt.wantSubCommands) {
				t.Errorf("Config.findSubCommand() SubCommands\ngot =%+v\nwant=%+v", gotSubCommands, tt.wantSubCommands)
			}
			if fmt.Sprintf("%+v", gotArgsWithoutCommand) != fmt.Sprintf("%+v", tt.wantArgsWithoutCommand) {
				t.Errorf("Config.findSubCommand() ArgsWithoutCommand\ngot =%+v\nwant=%+v", gotArgsWithoutCommand, tt.wantArgsWithoutCommand)
			}
		})
	}
}

func TestConfig_initParams(t *testing.T) {
	p1, _ := param.New(
		"P1",
		func(s string) error { return nil },
	)
	c1, _ := New(WithParams(*p1))
	_ = c1
	p211, _ := param.New(
		"P211",
		func(s string) error { return nil },
	)
	p21, _ := param.New(
		"P21",
		func(s string) error { return nil },
	)
	p22, _ := param.New(
		"P22",
		func(s string) error { return nil },
		param.WithIsSubCommandLocal(true),
	)
	p2, _ := param.New(
		"P2",
		func(s string) error { return nil },
	)
	c211, _ := New(WithParams(*p211))
	c21, _ := New(WithParams(*p21, *p22), WithSubCommand("sub211", c211))
	c2, _ := New(WithParams(*p2), WithSubCommand("sub21", c21))
	_ = c2

	c31, _ := New(WithParams(*p21, *p22))
	c3, _ := New(WithParams(*p2), WithSubCommand("sub31", c31))
	_ = c3

	type args struct {
		subCmd       []subcommand.SubCommand
		subCmdConfig *Manager
	}
	tests := []struct {
		name    string
		args    args
		wantPI  []paramname.ParamName
		wantErr bool
	}{
		{
			name:   "no sub cmd",
			args:   args{subCmd: []subcommand.SubCommand{""}, subCmdConfig: c1},
			wantPI: []paramname.ParamName{"P1"},
		},
		{
			name:   "with sub cmd",
			args:   args{subCmd: []subcommand.SubCommand{"", "sub21", "sub211"}, subCmdConfig: c2},
			wantPI: []paramname.ParamName{"P211", "P21", "P2"}, //No P22
		},
		{
			name:   "with sub cmd and local param",
			args:   args{subCmd: []subcommand.SubCommand{"", "sub31"}, subCmdConfig: c3},
			wantPI: []paramname.ParamName{"P22", "P21", "P2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Manager{}
			got, _, _, err := c.initParams(context.Background(), tt.args.subCmd, tt.args.subCmdConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.initParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.wantPI) {
				t.Errorf("Config.initParams() checking keys\ngot =%v\nwant=%v", got, tt.wantPI)
			}
			for _, k := range tt.wantPI {
				if _, ok := got[k]; !ok {
					t.Errorf("Config.initParams() checking keys\ngot =%v\nwant=%v", got, tt.wantPI)
				}
			}
		})
	}
}

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
			err:    errors.ParamConfigError{ParamName: pCity.Name, Err: fmt.Errorf("err desc")},
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
			err:    errors.ConfigError{Err: fmt.Errorf("err desc")},
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
			err:    errors.ConfigError{Err: fmt.Errorf("err desc")},
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
