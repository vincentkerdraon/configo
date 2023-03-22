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

func (rs *RotatingSecret) Deserialize(s string) error {
	secrets := strings.Split(s, ",")
	if len(secrets) != 3 {
		return fmt.Errorf("not 3 parts RotatingSecret as string, nothing to Deserialize")
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
