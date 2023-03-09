package lambdatype

import (
	"fmt"

	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/steptype"
)

// EventInput is what we receive from the secret manager.
type EventInput struct {
	//Step is the rotation step
	Step steptype.StepType
	//SecretARN is the secret ARN, example arn:aws:secretsmanager:us-east-1:388185734353:secret:testVincent-SdHK8l
	SecretARN string `json:"SecretId"`
	//VersionID is different from CURRENT or PENDING.
	VersionID string `json:"ClientRequestToken"`
}

func (e EventInput) Validate() error {
	if err := e.Step.Validate(); err != nil {
		return err
	}
	if e.SecretARN == "" || e.VersionID == "" {
		return fmt.Errorf("mandatory field SecretARN and VersionID")
	}
	return nil
}
