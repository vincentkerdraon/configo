package lambdaconf

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/lambdaconf/constraint"
)

type (
	LambdaConf struct {
		//By secretARN
		Secrets map[string]JSONSecretRotationConf
	}
	JSONSecretRotationConf struct {
		//By JSON key
		Keys map[string]JSONSecretRotationKeyConf
	}
	JSONSecretRotationKeyConf struct {
		Constraint     constraint.Constraint
		Prefix         string `json:",omitempty"`
		WithTime       bool   `json:",omitempty"`
		AlphaNumLength uint8
	}
)

func (c LambdaConf) Validate() error {
	for s, secrets := range c.Secrets {
		if s == "" {
			return fmt.Errorf("conf has empty SecretARN")
		}
		for k, keyConf := range secrets.Keys {
			if k == "" {
				return fmt.Errorf("conf has empty keys for SecretARN=%q", s)
			}
			if keyConf.AlphaNumLength < 8 {
				return fmt.Errorf("min for AlphaNumLength is 8 for SecretARN=%q, Key=%q", s, k)
			}
		}
	}
	return nil
}

func PrepareNewSecretFormatted(now time.Time, lambdaConf LambdaConf) func(ctx context.Context, secretARN string, secretOld string) (secretNew string, _ error) {
	return func(ctx context.Context, secretARN string, secretOld string) (secretNew string, _ error) {
		secretConf, f := lambdaConf.Secrets[secretARN]
		if !f {
			return "", fmt.Errorf("fail find SecretARN=%q in configuration", secretARN)
		}

		var decoded map[string]string
		err := json.Unmarshal([]byte(secretOld), &decoded)
		if err != nil {
			return "", fmt.Errorf("fail JSON decode old SecretARN=%q", secretARN)
		}

		for k, keyConf := range secretConf.Keys {
			switch keyConf.Constraint {
			case constraint.ConstraintNoOp:
			case constraint.ConstraintAlphaNum:
				secretNew := RandStringBytesRmndr(lettersAlphaNum, int(keyConf.AlphaNumLength))
				if keyConf.WithTime {
					secretNew = now.UTC().Format("20060102150405") + "-" + secretNew
				}
				if keyConf.Prefix != "" {
					secretNew = keyConf.Prefix + "-" + secretNew
				}
				decoded[k] = secretNew
			default:
				return "", fmt.Errorf("unknown Constraint=%q for SecretARN=%q", keyConf.Constraint, secretARN)
			}
		}

		encoded, err := json.Marshal(decoded)
		if err != nil {
			return "", fmt.Errorf("fail JSON encode new SecretARN=%q", secretARN)
		}
		return string(encoded), nil
	}
}
