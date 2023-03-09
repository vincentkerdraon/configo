package steptype

import (
	"fmt"
)

// StepType are the different lambda calls from the secret manager
//
// from github.com/aws/aws-sdk-go@v1.44.216/service/secretsmanager/api.go
// because stupid AWS lib won't export it
type StepType string

const (
	CreateSecret StepType = "createSecret"
	SetSecret    StepType = "setSecret"
	TestSecret   StepType = "testSecret"
	FinishSecret StepType = "finishSecret"
)

func (t StepType) String() string {
	return string(t)
}

func (t StepType) Validate() error {
	switch StepType(t) {
	case CreateSecret, SetSecret, TestSecret, FinishSecret:
		return nil
	default:
		return fmt.Errorf("unexpected StepType=%q", t)
	}
}
