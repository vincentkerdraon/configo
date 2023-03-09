package constraint

import (
	"fmt"
)

// Constraint gives the implementation to use for the secret rotation
type Constraint string

const (
	ConstraintAlphaNum Constraint = "AlphaNum"
	ConstraintNoOp     Constraint = "NoOp"
)

func (t Constraint) String() string {
	return string(t)
}
func (t *Constraint) Set(s string) error {
	switch Constraint(s) {
	case ConstraintAlphaNum, ConstraintNoOp:
		stepType := Constraint(s)
		*t = stepType
		return nil
	default:
		return fmt.Errorf("unexpected Constraint=%q", s)
	}
}
