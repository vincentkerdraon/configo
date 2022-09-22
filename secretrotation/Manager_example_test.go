package secretrotation_test

import (
	"errors"
	"fmt"

	"github.com/vincentkerdraon/configo/secretrotation"
)

func Example() {
	m := secretrotation.New()

	//Init not done
	_, err := m.Current()
	if !errors.Is(err, secretrotation.MissingInitValues{}) {
		panic(err)
	}

	//Provide input
	rs := secretrotation.NewRotatingSecret("my_secretA", "my_secretB", "my_secretC")
	err = m.Set(rs)
	handleErr(err)

	//Check received secret is allowed (this is the producer point of view)
	if !m.Allowed("my_secretC") {
		panic("expected secret OK")
	}

	//Get secret to use (this is the consumer point of view)
	secret, err := m.Current()
	handleErr(err)

	fmt.Print(secret)
	// Output:
	// my_secretB
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
