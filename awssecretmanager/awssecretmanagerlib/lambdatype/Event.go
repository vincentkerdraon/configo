package lambdatype

import (
	"fmt"

	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/steptype"
)

type EventInput struct {
	//Step is the rotation step
	Step steptype.StepType
	//SecretARN is the secret ARN, example arn:aws:secretsmanager:us-east-1:388185734353:secret:testVincent-SdHK8l
	SecretARN string `json:"SecretId"`
	//VersionID is different from CURRENT or PENDING.
	VersionID string `json:"ClientRequestToken"`
}

func (e EventInput) Validate() error {
	var step steptype.StepType

	//ok a bit weird but Set(s string) error already exists and is convenient
	if err := step.Set(e.Step.String()); err != nil {
		return err
	}
	if e.SecretARN == "" || e.VersionID == "" {
		return fmt.Errorf("mandatory field SecretARN and VersionID")
	}
	return nil
}
