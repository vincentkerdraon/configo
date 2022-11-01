package secretrotation

import "fmt"

type MissingInitValuesError struct{}

func (MissingInitValuesError) Error() string {
	return "Missing secrets init values"
}

type InvalidSecretError struct {
	Err error
}

func (err InvalidSecretError) Error() string {
	return fmt.Sprintf("Invalid secret, %s", err.Err)
}

func (e InvalidSecretError) Unwrap() error { return e.Err }
