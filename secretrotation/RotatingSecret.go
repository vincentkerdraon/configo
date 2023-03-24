package secretrotation

import (
	"fmt"
	"strings"
)

type (
	RotatingSecret struct {
		Previous Secret
		Current  Secret
		Pending  Secret
	}
)

func NewRotatingSecret(previous, current, pending Secret) RotatingSecret {
	return RotatingSecret{
		Previous: previous,
		Current:  current,
		Pending:  pending,
	}
}

func (rs RotatingSecret) Validate() error {
	if err := rs.Previous.Validate(); err != nil {
		return err
	}
	if err := rs.Current.Validate(); err != nil {
		return err
	}
	if err := rs.Pending.Validate(); err != nil {
		return err
	}
	return nil
}

func (rs RotatingSecret) Serialize() string {
	return fmt.Sprintf("%s,%s,%s", rs.Previous, rs.Current, rs.Pending)
}

func (rs *RotatingSecret) Set(s string) error {
	return rs.Deserialize(s)
}

// Deserialize will populate the RotatingSecret object based on the string value
//
// If string empty => error.
// If 1 part string => all 3 values of the secret will be the same.
// If 3 part string, comma separated => set into RotatingSecret.
// Else => error
func (rs *RotatingSecret) Deserialize(s string) error {
	if len(s) == 0 {
		return fmt.Errorf("fail Deserialize empty string as RotatingSecret")
	}
	secrets := strings.Split(s, ",")
	if len(secrets) == 1 {
		rs.Previous = Secret(secrets[0])
		rs.Current = Secret(secrets[0])
		rs.Pending = Secret(secrets[0])
		if err := rs.Validate(); err != nil {
			return err
		}
		return nil
	}
	if len(secrets) != 3 {
		return fmt.Errorf("not 3 parts RotatingSecret as string, fail Deserialize")
	}
	rs.Previous = Secret(secrets[0])
	rs.Current = Secret(secrets[1])
	rs.Pending = Secret(secrets[2])
	if err := rs.Validate(); err != nil {
		return err
	}
	return nil
}

// Range iterates over the secrets
func (rs RotatingSecret) Range(f func(Secret) (continueRange bool)) {
	for _, s := range []Secret{rs.Current, rs.Pending, rs.Previous} {
		if !f(s) {
			return
		}
	}
}

func (rs RotatingSecret) RedactSecret(in string) string {
	rs.Range(func(s Secret) (continueRange bool) {
		in = s.RedactSecret(in)
		return true
	})
	return in
}
