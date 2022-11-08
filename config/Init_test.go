package config

import (
	"context"
	"fmt"
	"testing"

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
		// {
		// 	name:   "no sub cmd",
		// 	args:   args{subCmd: []subcommand.SubCommand{}, subCmdConfig: c1},
		// 	wantPI: []paramname.ParamName{"P1"},
		// },
		{
			name:   "with sub cmd",
			args:   args{subCmd: []subcommand.SubCommand{"sub21", "sub211"}, subCmdConfig: c2},
			wantPI: []paramname.ParamName{"P211", "P21", "P2"}, //No P22
		},
		// {
		// 	name:   "with sub cmd and local param",
		// 	args:   args{subCmd: []subcommand.SubCommand{"sub31"}, subCmdConfig: c3},
		// 	wantPI: []paramname.ParamName{"P22", "P21", "P2"},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Manager{}
			got, _, _, _, err := c.initParams(context.Background(), []subcommand.SubCommand{subCommandLevel0}, tt.args.subCmd, tt.args.subCmdConfig)
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
