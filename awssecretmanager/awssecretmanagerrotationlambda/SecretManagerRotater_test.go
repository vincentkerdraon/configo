package awssecretmanagerrotationlambda

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/lambdaconf"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/lambdatype"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/steptype"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/versionstage"
)

type awsSecretsManagerMock struct{}

func (sm awsSecretsManagerMock) GetRandomPasswordWithContext(ctx context.Context, input *secretsmanager.GetRandomPasswordInput, opts ...request.Option) (*secretsmanager.GetRandomPasswordOutput, error) {
	return &secretsmanager.GetRandomPasswordOutput{
		RandomPassword: aws.String("RandomPassword1"),
	}, nil
}
func (sm awsSecretsManagerMock) DescribeSecret(input *secretsmanager.DescribeSecretInput) (*secretsmanager.DescribeSecretOutput, error) {
	return &secretsmanager.DescribeSecretOutput{
		RotationEnabled: aws.Bool(true),
		VersionIdsToStages: map[string][]*string{
			"versionID1":                  nil,
			versionstage.Pending.String(): nil,
		},
	}, nil
}
func (sm awsSecretsManagerMock) GetSecretValueWithContext(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error) {
	return &secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(`{"key1":"val1","key2":"val2"}`),
	}, nil
}
func (sm awsSecretsManagerMock) PutSecretValueWithContext(ctx aws.Context, input *secretsmanager.PutSecretValueInput, opts ...request.Option) (*secretsmanager.PutSecretValueOutput, error) {
	return &secretsmanager.PutSecretValueOutput{}, nil
}
func (sm awsSecretsManagerMock) UpdateSecretVersionStageWithContext(ctx context.Context, input *secretsmanager.UpdateSecretVersionStageInput, opts ...request.Option) (*secretsmanager.UpdateSecretVersionStageOutput, error) {
	return &secretsmanager.UpdateSecretVersionStageOutput{}, nil
}

func TestSecretManagerRotater(t *testing.T) {
	now, err := time.Parse(time.DateTime, "2023-03-08 15:04:05")
	if err != nil {
		t.Fatal(err)
	}
	lambdaConf := lambdaconf.LambdaConf{}
	if err := lambdaConf.Validate(); err != nil {
		t.Fatal(err)
	}
	svc := awsSecretsManagerMock{}
	prepareSecretNew := lambdaconf.PrepareNewSecretFormatted(now, lambdaConf)
	smr := New(
		WithAWSSecretsManager(svc),
		WithPrepareSecret(prepareSecretNew),
	)
	err = smr.HandleRequest(context.Background(), lambdatype.EventInput{
		Step:      steptype.CreateSecret,
		SecretARN: "secretARN1",
		VersionID: "versionID1",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = smr.HandleRequest(context.Background(), lambdatype.EventInput{
		Step:      steptype.SetSecret,
		SecretARN: "secretARN1",
		VersionID: "versionID1",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = smr.HandleRequest(context.Background(), lambdatype.EventInput{
		Step:      steptype.TestSecret,
		SecretARN: "secretARN1",
		VersionID: "versionID1",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = smr.HandleRequest(context.Background(), lambdatype.EventInput{
		Step:      steptype.FinishSecret,
		SecretARN: "secretARN1",
		VersionID: "versionID1",
	})
	if err != nil {
		t.Fatal(err)
	}
}
