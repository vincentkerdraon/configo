package secretrotation

import (
	"fmt"
	"strings"
)

type (
	Secret string
)

const SecretRedacted = "[redacted]"

func (s Secret) Validate() error {
	if s == "" {
		return fmt.Errorf("empty secret")
	}
	return nil
}
func (s Secret) String() string {
	return string(s)
}

func (s *Secret) Set(in string) error {
	*s = Secret(in)
	return nil
}

func (s Secret) RedactSecret(in string) string {
	return strings.ReplaceAll(in, s.String(), SecretRedacted)
}
