package param

import (
	"fmt"
	"reflect"
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
		name  paramname.ParamName
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

type interfaceWithSetter struct {
	val       string
	unchanged string
}

func (i *interfaceWithSetter) Set(s string) error {
	if s == "err" {
		return fmt.Errorf("err")
	}
	i.val = s
	return nil
}

func TestNewParamFromStructTag_parse(t *testing.T) {
	type SuperString string
	type struct1 struct {
		KeySS      SuperString
		KeyInt     int
		KeyUint8   uint8
		KeyFloat32 float32
		KeyBool    bool
		ISetter    interfaceWithSetter
		ISetterP   *interfaceWithSetter
	}

	myStruct1 := &struct1{
		ISetter: interfaceWithSetter{
			unchanged: "unchanged",
		},
		ISetterP: &interfaceWithSetter{
			unchanged: "unchanged",
		},
	}
	_ = myStruct1

	type args struct {
		i     *struct1
		name  paramname.ParamName
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
				i:     myStruct1,
				name:  "KeySS",
				parse: func(s string) error { myStruct1.KeySS = SuperString(fmt.Sprintf("_%s_", s)); return nil },
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
				i:    myStruct1,
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
				i:    myStruct1,
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
				i:    myStruct1,
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
				i:    myStruct1,
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
				i:    myStruct1,
				name: "KeyBool",
			},
			val: "true",
			check: func(t *testing.T, i *struct1) {
				if i.KeyBool != true {
					t.Errorf("got =%t\n", i.KeyBool)
				}
			},
		},
		{
			name: "keep initial value if not overridded",
			args: args{
				i:    &struct1{KeySS: "initial", KeyInt: 12},
				name: "KeySS",
			},
			val: "",
			check: func(t *testing.T, i *struct1) {
				if i.KeySS != "initial" {
					t.Errorf("got =%s\n", i.KeySS)
				}
			},
		},
		{
			name: "interfaceWithSetterP",
			args: args{
				name: "ISetterP",
				i:    myStruct1,
			},
			val: "value",
			check: func(t *testing.T, i *struct1) {
				if i.ISetterP.unchanged != "unchanged" {
					t.Errorf("lost sister value in struct\n")
				}
				if i.ISetterP.val != "value" {
					t.Errorf("got =%s\n", i.ISetter.val)
				}
			},
		},
		// -- Won't work.
		// {
		// 	name: "interfaceWithSetter",
		// 	args: args{
		// 		name: "ISetter",
		// 		i:    myStruct1,
		// 	},
		// 	val: "value",
		// 	check: func(t *testing.T, i *struct1) {
		// 		if i.ISetter.unchanged != "unchanged" {
		// 			t.Errorf("lost sister value in struct\n")
		// 		}
		// 		if i.ISetter.val != "value" {
		// 			t.Errorf("got =%s\n", i.ISetter.val)
		// 		}
		// 	},
		// },
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

func TestIterateStructFields(t *testing.T) {
	type Empty struct{}
	type NonExported struct {
		nonExported bool
	}
	type Simple struct {
		A string
		B uint64
	}
	type Complex struct {
		Empty
		D Simple
		Simple
		C string
		NonExported
	}

	type args struct {
		v interface{}
		f func(name paramname.ParamName) error
	}
	namesReceived := []paramname.ParamName{}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		namesExpected []paramname.ParamName
	}{
		{
			name: "Empty",
			args: args{
				v: &Empty{},
				f: func(name paramname.ParamName) error { t.Fatal(); return nil },
			},
			namesExpected: []paramname.ParamName{},
		},
		{
			name: "NonExported",
			args: args{
				v: &NonExported{},
				f: func(name paramname.ParamName) error { t.Fatal(); return nil },
			},
			namesExpected: []paramname.ParamName{},
		},
		{
			name: "error",
			args: args{
				v: &Simple{},
				f: func(name paramname.ParamName) error {
					return fmt.Errorf("err")
				},
			},
			wantErr:       true,
			namesExpected: []paramname.ParamName{},
		},
		{
			name: "Simple",
			args: args{
				v: &Simple{},
				f: func(name paramname.ParamName) error {
					namesReceived = append(namesReceived, name)
					return nil
				},
			},
			namesExpected: []paramname.ParamName{"A", "B"},
		},
		{
			name: "Complex",
			args: args{
				v: &Complex{},
				f: func(name paramname.ParamName) error {
					namesReceived = append(namesReceived, name)
					return nil
				},
			},
			// namesExpected: []string{"Empty", "D", "Simple", "C", "NonExported"},
			namesExpected: []paramname.ParamName{"C"},
		},
	}
	for _, tt := range tests {
		namesReceived = []paramname.ParamName{}
		t.Run(tt.name, func(t *testing.T) {
			if err := IterateStructFields(tt.args.v, tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("IterateStructFields() error = %v, wantErr %v", err, tt.wantErr)
			}
			// fmt.Printf("\ngot =%T %#v %q\nwant=%T %#v %q", namesReceived, namesReceived, namesReceived, tt.namesExpected, tt.namesExpected, tt.namesExpected)
			if !reflect.DeepEqual(namesReceived, tt.namesExpected) {
				t.Errorf("IterateStructFields()\ngot =%v\nwant=%v", namesReceived, tt.namesExpected)
			}
		})
	}
}
