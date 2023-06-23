package secretrotation

import (
	"crypto/subtle"
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

// IsAllowed will check if a given key matches the secrets
func (s Secret) IsAllowed(key Secret) bool {
	//using a constant time comparison.
	//It will always take the same time when wrong, being closer to the solution or not.
	return 1 == subtle.ConstantTimeCompare([]byte(key), []byte(s))
}
