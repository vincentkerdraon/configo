package config

import (
	"context"
	"testing"
)

func Test_param_default_value(t *testing.T) {
	type City string
	user := struct {
		Name string `default:"NameDefault"`
	}{
		Name: "NameInit",
	}

	//Read the tags on the struct field. And tries to match simple types in the automatic parse() function.
	c, err := New(WithParamsFromStructTag(&user, ""))
	if err != nil {
		t.Fatal(err)
	}
	err = c.Init(
		context.Background(),
		//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
		WithInputArgs([]string{}),
	)
	if err != nil {
		t.Fatal(err)
	}

	//We could expect NameInit, because there was already a value so setting the default should not apply.
	//But configo does not know about the current content. it will only check the inputs.
	expected := "NameDefault"
	if user.Name != expected {
		t.Errorf("use init struct value\ngot =%q\nwant=%q", user.Name, expected)
	}
}
