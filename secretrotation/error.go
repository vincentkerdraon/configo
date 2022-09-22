package secretrotation

import "fmt"

type MissingInitValues struct{}

func (MissingInitValues) Error() string {
	return "Missing secrets init values"
}

type InvalidSecret struct {
	Err error
}

func (err InvalidSecret) Error() string {
	return fmt.Sprintf("Invalid secret, %s", err.Err)
}
