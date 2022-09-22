package param

import (
	"fmt"
	"testing"

	"github.com/vincentkerdraon/configo/config/param/paramname"
)

func TestNewParamFromStructTag(t *testing.T) {
	type struct1 struct {
		Key1 string `flag:"specialKey" envVar:"THE_KEY" mandatory:"true" desc:"bla bla desc" examples:"aa;bb" default:"aa" exclusiveTags:"ExampleS" enumValues:"aa;bb;cc;dd"`
	}

	parseNoop := func(s string) error { return nil }

	type args struct {
		i     struct1
		name  string
		parse func(s string) error
	}
	tests := []struct {
		name  string
		args  args
		want  *Param
		check func(i struct1, err error) error
	}{
		{
			name: "ok",
			args: args{
				i:     struct1{},
				name:  "Key1",
				parse: parseNoop,
			},
			want: &Param{
				Name:        "Key1",
				Parse:       parseNoop,
				Flag:        Flag{Use: true, Name: "specialKey"},
				EnvVar:      EnvVar{Use: true, Name: "THE_KEY"},
				IsMandatory: true,
				Desc:        "bla bla desc",
				Examples:    []string{"aa", "bb"},
				Exclusive:   []paramname.ParamName{"ExampleS"},
				EnumValues:  []string{"aa", "bb", "cc", "dd"},
				Default:     "aa",
			},
			check: func(i struct1, err error) error {
				if err != nil {
					return err
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewParamFromStructTag(&tt.args.i, tt.args.name, tt.args.parse)
			//check Parse() in a different test
			got.Parse = nil
			tt.want.Parse = nil

			if fmt.Sprintf("%+v", got) != fmt.Sprintf("%+v", tt.want) {
				t.Errorf("NewParamFromStructTag() \ngot =%+v\nwant=%+v", got, tt.want)
			}
			if errCheck := tt.check(tt.args.i, err); errCheck != nil {
				t.Errorf("NewParamFromStructTag() errCheck: %s", errCheck)
			}
		})
	}
}

func TestNewParamFromStructTag_parse(t *testing.T) {
	type SuperString string
	type struct1 struct {
		KeySS      SuperString
		KeyInt     int
		KeyUint8   uint8
		KeyFloat32 float32
		KeyBool    bool
	}

	myStruct := &struct1{}

	type args struct {
		i     *struct1
		name  string
		parse func(s string) error
	}
	tests := []struct {
		name    string
		args    args
		val     string
		wantErr bool
		check   func(t *testing.T, i *struct1)
	}{
		{
			name: "custom parse provided",
			args: args{
				i:     myStruct,
				name:  "KeySS",
				parse: func(s string) error { myStruct.KeySS = SuperString(fmt.Sprintf("_%s_", s)); return nil },
			},
			val: "customParse",
			check: func(t *testing.T, i *struct1) {
				if i.KeySS != "_customParse_" {
					t.Errorf("got =%s\n", i.KeySS)
				}
			},
		},
		{
			name: "parse auto: custom type string",
			args: args{
				i:    myStruct,
				name: "KeySS",
			},
			val: "valueS",
			check: func(t *testing.T, i *struct1) {
				if i.KeySS != "valueS" {
					t.Errorf("got =%s\n", i.KeySS)
				}
			},
		},
		{
			name: "parse auto: int",
			args: args{
				i:    myStruct,
				name: "KeyInt",
			},
			val: "12",
			check: func(t *testing.T, i *struct1) {
				if i.KeyInt != 12 {
					t.Errorf("got =%d\n", i.KeyInt)
				}
			},
		},
		{
			name: "parse auto: uint8",
			args: args{
				i:    myStruct,
				name: "KeyUint8",
			},
			val: "12",
			check: func(t *testing.T, i *struct1) {
				if i.KeyUint8 != 12 {
					t.Errorf("got =%d\n", i.KeyUint8)
				}
			},
		},
		{
			name: "parse auto: float32",
			args: args{
				i:    myStruct,
				name: "KeyFloat32",
			},
			val: "12.12",
			check: func(t *testing.T, i *struct1) {
				if i.KeyFloat32 != 12.12 {
					t.Errorf("got =%f\n", i.KeyFloat32)
				}
			},
		},
		{
			name: "parse auto: bool",
			args: args{
				i:    myStruct,
				name: "KeyBool",
			},
			val: "true",
			check: func(t *testing.T, i *struct1) {
				if i.KeyBool != true {
					t.Errorf("got =%t\n", i.KeyBool)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParamFromStructTag(tt.args.i, tt.args.name, tt.args.parse)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewParamFromStructTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := p.Parse(tt.val); err != nil {
				t.Errorf("Parse() error = %v", err)
			}
			tt.check(t, tt.args.i)
		})
	}
}
