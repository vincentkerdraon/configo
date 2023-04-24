package param

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/vincentkerdraon/configo/config/errors"
	"github.com/vincentkerdraon/configo/config/param/paramname"
)

// StructTag* are the expected tags to decode for automatic configuration detection
const (
	StructTagFlag          = "flag"
	StructTagEnvVar        = "envVar"
	StructTagMandatory     = "mandatory"
	StructTagDesc          = "desc"
	StructTagDefault       = "default"
	StructTagExamples      = "examples"
	StructTagExclusiveTags = "exclusiveTags"
	StructTagEnumValues    = "enumValues"
)

// literalStore tries to set a value into a generic type. Best effort.
func literalStore(s string, v reflect.Value) error {
	//inspired by std json decode: func (d *decodeState) literalStore()
	//heavily modified.

	// use interface "Set(string) error" if defined. See std flag lib for flag.Var()
	if v.CanInterface() {
		type setter interface {
			Set(string) error
		}

		if setter, ok := v.Interface().(setter); ok {
			return setter.Set(s)
		}

		//TODO it would be great if we could detect non-pointer struct fulfilling the interface. I am giving up for now.
	}

	switch v.Type().String() {
	case "time.Duration":
		d, err := time.ParseDuration(s)
		if err != nil {
			return &json.UnmarshalTypeError{Value: fmt.Sprintf("duration:%q", s), Type: v.Type()}
		}
		v.SetInt(d.Nanoseconds())
		return nil
	}

	switch v.Kind() {
	default:
		val := reflect.ValueOf(s)
		v.Set(val.Convert(v.Type()))
	case reflect.Interface:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return &json.UnmarshalTypeError{Value: fmt.Sprintf("number:%q", s), Type: reflect.TypeOf(0.0)}
		}
		if v.NumMethod() != 0 {
			return &json.UnmarshalTypeError{Value: fmt.Sprintf("number:%q", s), Type: v.Type()}
		}
		v.Set(reflect.ValueOf(f))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || v.OverflowInt(n) {
			return &json.UnmarshalTypeError{Value: fmt.Sprintf("number:%q", s), Type: v.Type()}
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil || v.OverflowUint(n) {
			return &json.UnmarshalTypeError{Value: fmt.Sprintf("number:%q", s), Type: v.Type()}
		}
		v.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil || v.OverflowFloat(n) {
			return &json.UnmarshalTypeError{Value: fmt.Sprintf("number:%q", s), Type: v.Type()}
		}
		v.SetFloat(n)

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return &json.UnmarshalTypeError{Value: fmt.Sprintf("bool:%q", s), Type: v.Type()}
		}
		v.SetBool(b)
	}

	return nil
}

// NewParamFromStructTag tries to automatically define a param using reflection on the struct.
// If parse is null, it tries to do the matching (best effort).
func NewParamFromStructTag(
	v interface{},
	name paramname.ParamName,
	parse func(s string) error,
	opts ...paramOption,
) (*Param, error) {
	paramName := paramname.ParamName(name)
	field, ok := reflect.TypeOf(v).Elem().FieldByName(name.String())
	if !ok {
		return nil, errors.ParamConfigError{ParamName: paramName, Err: fmt.Errorf("fail find struct field:%q", name)}
	}

	if parse == nil && field.IsExported() {
		parse = func(s string) error {
			if s == "" {
				return nil
			}

			structFieldValue := reflect.ValueOf(v).Elem().FieldByName(name.String())
			if err := literalStore(s, structFieldValue); err != nil {
				return errors.ParamConfigError{ParamName: paramName, Err: err}
			}
			return nil
		}
	}

	paramOptions := []paramOption{}

	if alias, ok := field.Tag.Lookup(StructTagFlag); ok {
		if alias == "-" {
			paramOptions = append(paramOptions, WithFlag(WithReadFlag(false)))
		} else if alias != "" {
			paramOptions = append(paramOptions, WithFlag(WithFlagName(alias)))
		}
	}

	if alias, ok := field.Tag.Lookup(StructTagEnvVar); ok {
		if alias == "-" {
			paramOptions = append(paramOptions, WithEnvVar(WithReadEnvVar(false)))
		} else if alias != "" {
			paramOptions = append(paramOptions, WithEnvVar(WithEnvVarName(alias)))
		}
	}

	if alias, ok := field.Tag.Lookup(StructTagMandatory); ok {
		b, err := strconv.ParseBool(alias)
		if err != nil {
			return nil, errors.ParamConfigError{ParamName: paramName, Err: fmt.Errorf("struct tag:%q value must be boolean", StructTagMandatory)}
		}
		paramOptions = append(paramOptions, WithIsMandatory(b))
	}

	if alias, ok := field.Tag.Lookup(StructTagDesc); ok {
		paramOptions = append(paramOptions, WithDesc(alias))
	}

	if alias, ok := field.Tag.Lookup(StructTagDefault); ok {
		paramOptions = append(paramOptions, WithDefault(alias))
	}

	if alias, ok := field.Tag.Lookup(StructTagExamples); ok {
		paramOptions = append(paramOptions, WithExamples(strings.Split(alias, ";")...))
	}

	if alias, ok := field.Tag.Lookup(StructTagExclusiveTags); ok {
		exclusiveParams := []paramname.ParamName{}
		for _, et := range strings.Split(alias, ";") {
			exclusiveParams = append(exclusiveParams, paramname.ParamName(et))
		}
		paramOptions = append(paramOptions, WithExclusive(exclusiveParams...))
	}

	if alias, ok := field.Tag.Lookup(StructTagEnumValues); ok {
		paramOptions = append(paramOptions, WithEnumValues(strings.Split(alias, ";")...))
	}

	return New(paramname.ParamName(field.Name), parse, append(paramOptions, opts...)...)
}

// ParamsFromStructTag reads the struct tags, using all the default options otherwise.
// A prefix can be added in front of the param name + flag name + env var name
func ParamsFromStructTag(
	v interface{},
	prefix string,
) ([]*Param, error) {
	params := []*Param{}
	err := IterateStructFields(v, func(name paramname.ParamName) error {
		p, err := NewParamFromStructTag(v, name, nil, WithPrefix(prefix))
		if err != nil {
			return err
		}
		params = append(params, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return params, nil
}

// IterateStructFields finds all the exported fields in a *struct.
//
// Input MUST be a pointer to the struct. To avoid `reflect: Elem of invalid type`.
// This ignores all the struct fields. You have to make explicit.
// This is mostly a helper to call NewParamFromStructTag on every fields with some added logic like a name prefix.
func IterateStructFields(
	v interface{},
	f func(name paramname.ParamName) error,
) error {
	typeOfV := reflect.TypeOf(v)
	if typeOfV.Kind() != reflect.Ptr {
		//panic instead of returning an error here,
		//easier to find what the faulty struct.
		panic(fmt.Errorf("expect Ptr, got %s", typeOfV.Kind()))
	}

	//Return nice error in case of double pointer
	// if reflect.TypeOf(typeOfV).Kind() == reflect.Ptr {
	// 	return fmt.Errorf("expect Ptr, got double pointer")
	// }

	st := typeOfV.Elem()
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		//ignore embedded struct
		if field.Type.Kind() == reflect.Struct {
			continue
		}
		if !field.IsExported() {
			continue
		}
		if err := f(paramname.ParamName(field.Name)); err != nil {
			return err
		}
	}
	return nil
}
