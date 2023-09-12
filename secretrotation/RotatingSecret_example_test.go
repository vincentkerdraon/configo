package secretrotation_test

import (
	"fmt"

	"github.com/vincentkerdraon/configo/secretrotation"
)

func ExampleRotatingSecret_Validate() {
	rs := secretrotation.NewRotatingSecret("my_secretA", "my_secretB", "my_secretC")
	if rs.Validate() != nil {
		panic("expect ok")
	}

	rs = secretrotation.NewRotatingSecret("my_secretA", "my_secretB", "")
	if rs.Validate() == nil {
		panic("expect fail")
	}

	fmt.Printf("Validate=%q\n", rs.Validate())
	// Output:
	// Validate="empty secret"
}

func ExampleRotatingSecret_Serialize() {
	rs := secretrotation.NewRotatingSecret("my_secretA", "my_secretB", "my_secretC")
	serial := rs.Serialize()

	//Serialize
	rs2 := secretrotation.RotatingSecret{}

	if err := rs2.Deserialize(serial); err != nil {
		panic("expect ok")
	}
	if rs2.Current != rs.Current {
		panic("expect ok")
	}

	fmt.Printf("serial=%q, rs2=%+v\n", serial, rs2)
	// Output:
	// serial="my_secretA,my_secretB,my_secretC", rs2={Previous:my_secretA Current:my_secretB Pending:my_secretC}
}

func ExampleRotatingSecret_RedactSecret() {
	rs := secretrotation.NewRotatingSecret("my_secretA", "my_secretB", "my_secretC")
	const input = "my string including my_secretA"
	const expected = "my string including [redacted]"
	redacted := rs.RedactSecret(input)

	if redacted != expected {
		panic("expect ok")
	}

	fmt.Printf("redacted=%q\n", redacted)
	// Output:
	// redacted="my string including [redacted]"
}

func ExampleRotatingSecret_Allowed() {
	rs := secretrotation.NewRotatingSecret("my_secretA", "my_secretB", "my_secretC")
	// For example we receive a Shared secret in a http request, and we want to check it matches one of the known secrets
	// This function will always take the same time, in order to make it harder to detect how close an attack is to the solution by looking at the processing time.
	if !rs.Allowed("my_secretC") {
		panic("expect ok")
	}
}
