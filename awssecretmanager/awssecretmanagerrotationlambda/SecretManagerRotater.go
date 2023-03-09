package awssecretmanagerrotationlambda

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/lambdatype"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/steptype"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/versionstage"
)

type (
	AWSSecretsManager interface {
		GetRandomPasswordWithContext(ctx context.Context, input *secretsmanager.GetRandomPasswordInput, opts ...request.Option) (*secretsmanager.GetRandomPasswordOutput, error)
		DescribeSecret(input *secretsmanager.DescribeSecretInput) (*secretsmanager.DescribeSecretOutput, error)
		GetSecretValueWithContext(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error)
		PutSecretValueWithContext(ctx aws.Context, input *secretsmanager.PutSecretValueInput, opts ...request.Option) (*secretsmanager.PutSecretValueOutput, error)
		UpdateSecretVersionStageWithContext(ctx context.Context, input *secretsmanager.UpdateSecretVersionStageInput, opts ...request.Option) (*secretsmanager.UpdateSecretVersionStageOutput, error)
	}

	impl struct {
		svc           AWSSecretsManager
		logger        LeveledLogger
		prepareSecret func(ctx context.Context, secretARN string, secretOld string) (secretNew string, _ error)

		// setSecret should set the AWSPENDING secret in the service that the secret belongs to. For example, if the secret is a database
		// credential, this method should take the value of the AWSPENDING secret and set the user's password to this value in the database.
		setSecret func(ctx context.Context, secretARN string, versionID string) error

		// testSecret should validate that the AWSPENDING secret works in the service that the secret belongs to. For example, if the secret
		// is a database credential, this method should validate that the user can login with the password in AWSPENDING and that the user has
		// all of the expected permissions against the database.
		testSecret func(ctx context.Context, secretARN string, versionID string) error
	}
)

func New(opts ...Option) *impl {
	r := &impl{}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(r)
	}
	if r.logger == nil {
		r.logger = NewLeveledLoggerStandard(LogLevelInfo)
	}
	if r.svc == nil {
		r.svc = SvcSecretManagerDefault()
	}
	if r.prepareSecret == nil {
		r.prepareSecret = func(ctx context.Context, secretARN, secretOld string) (secretNew string, _ error) {
			new, err := r.svc.GetRandomPasswordWithContext(ctx, &secretsmanager.GetRandomPasswordInput{})
			if err != nil {
				return "", err
			}
			return new.String(), nil
		}
	}
	if r.setSecret == nil {
		r.setSecret = func(ctx context.Context, secretARN, versionID string) error {
			r.logger.Info("setSecret: no operation")
			return nil
		}
	}
	if r.testSecret == nil {
		r.testSecret = func(ctx context.Context, secretARN, versionID string) error {
			r.logger.Info("testSecret: no operation")
			return nil
		}
	}
	return r
}

// HandleRequest is special naming for Lambda.
//
// check possible signature in https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html .
// this is automatically using JSON Unmarshal.
//
// Example:
// [{Step:createSecret secretARN:arn:aws:secretsmanager:us-east-1:388185734353:secret:testVincent-SdHK8l VersionID:a22e23fc-6d02-45f0-845d-e0bad50838f0}]
func (r impl) HandleRequest(ctx context.Context, e lambdatype.EventInput) error {
	now := time.Now()
	r.logger.Debug("new event: %+v", e)
	defer r.logger.Trace("done in %s", time.Since(now))

	lc, _ := lambdacontext.FromContext(ctx)
	r.logger.Debug("lambda context:%+v", *lc)

	// Make sure
	// - the secret exists
	// - permissions are set
	// - version is staged correctly
	metadata, err := r.svc.DescribeSecret(
		&secretsmanager.DescribeSecretInput{
			SecretId: aws.String(e.SecretARN),
		})
	if err != nil {
		//Could also check ResourceNotFoundException for specific error
		return fmt.Errorf("fail describe SecretARN=%q, %w", e.SecretARN, err)
	}
	if metadata == nil {
		return fmt.Errorf("fail describe SecretARN=%q, no output", e.SecretARN)
	}
	if metadata.RotationEnabled == nil || !*metadata.RotationEnabled {
		return fmt.Errorf("secret rotation not configured for SecretARN=%q", e.SecretARN)
	}

	//versionID is used by AWS system to validate what stage of the rotation is happening.
	//if AWSCURRENT => nothing to do
	//Expecting that AWSPENDING is set.

	stages, f := metadata.VersionIdsToStages[e.VersionID]
	if !f {
		return fmt.Errorf("Secret VersionID=%q not in VersionIdsToStages for SecretARN=%q", e.VersionID, e.SecretARN)
	}

	foundCurrent := false
	foundPending := false
	for _, stage := range stages {
		if stage == nil {
			continue
		}
		switch *stage {
		case versionstage.Current.String():
			foundCurrent = true
		case versionstage.Pending.String():
			foundPending = true
		}
	}
	if foundCurrent {
		r.logger.Info("Secret already set to VersionStage=%q for VersionID=%q, SecretARN=%q", versionstage.Current, e.VersionID, e.SecretARN)
		return nil
	}
	if !foundPending {
		return fmt.Errorf("Secret should have VersionStage=%q for VersionID=%q, SecretARN=%q", versionstage.Pending, e.VersionID, e.SecretARN)
	}

	switch e.Step {
	case steptype.CreateSecret:
		err = r.createSecret(ctx, e.SecretARN, e.VersionID)
	case steptype.SetSecret:
		err = r.setSecret(ctx, e.SecretARN, e.VersionID)
	case steptype.TestSecret:
		err = r.testSecret(ctx, e.SecretARN, e.VersionID)
	case steptype.FinishSecret:
		err = r.finishSecret(ctx, e.SecretARN, e.VersionID, metadata)
	default:
		err = fmt.Errorf("unknown Step=%q", e.Step)
	}
	if err != nil {
		return fmt.Errorf("for SecretARN=%q, %w", e.SecretARN, err)
	}
	return nil
}

// createSecret first checks for the existence of a secret for the passed in token. If one does not exist, it will generate a
// new secret and put it with the passed in token.
func (r impl) createSecret(ctx context.Context, secretARN string, versionID string) error {
	// # Make sure the current secret exists
	old, err := r.svc.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretARN),
		VersionStage: aws.String(versionstage.Current.String()),
	})
	if err != nil {
		return fmt.Errorf("secret does not exist, VersionStage=%q, %w", versionstage.Current, err)
	}

	// # Now try to get the secret version, if that fails, put a new secret
	_, err = r.svc.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretARN),
		VersionStage: aws.String(versionstage.Pending.String()),
	})
	if err == nil {
		//already have a secret, nothing to do
		r.logger.Info("createSecret: already has VersionStage=%q, nothing to do, SecretARN=%q", versionstage.Pending, secretARN)
		return nil
	}
	if !strings.Contains(err.Error(), "ResourceNotFoundException") {
		return fmt.Errorf("fail GetSecretValue by VersionStage=%q, %w", versionstage.Pending, err)
	}

	secretNew, err := r.prepareSecret(ctx, secretARN, *old.SecretString)
	if err != nil {
		return fmt.Errorf("fail prepareSecret, %w", err)
	}

	// # Put the secret
	_, err = r.svc.PutSecretValueWithContext(ctx, &secretsmanager.PutSecretValueInput{
		SecretId:           aws.String(secretARN),
		ClientRequestToken: aws.String(versionID),
		VersionStages:      []*string{aws.String(versionstage.Pending.String())},
		SecretString:       aws.String(secretNew),
	})
	if err != nil {
		return fmt.Errorf("fail PutSecretValue, VersionStages=%q, %w", versionstage.Pending, err)
	}
	r.logger.Info("createSecret: Successfully put secret for VersionID=%q, VersionStage=%q, SecretARN=%q", versionID, versionstage.Pending, secretARN)
	return nil
}

// finishSecret finalizes the rotation process by marking the secret version passed in as the AWSCURRENT secret.
func (r impl) finishSecret(ctx context.Context, secretARN string, versionID string, metadata *secretsmanager.DescribeSecretOutput) error {
	// # First describe the secret to get the current version
	var versionCurrent string
F:
	for version, stages := range metadata.VersionIdsToStages {
		for _, s := range stages {
			if s != nil && *s == versionstage.Current.String() {
				if version == versionID {
					// # The correct version is already marked as current, return
					r.logger.Info("finishSecret: already has VersionStage=%q, nothing to do, VersionID=%q, SecretARN=%q", versionstage.Current, versionID, secretARN)
					return nil
				}
				versionCurrent = version
				break F
			}
		}
	}

	// # Finalize by staging the secret version current
	_, err := r.svc.UpdateSecretVersionStageWithContext(ctx, &secretsmanager.UpdateSecretVersionStageInput{
		SecretId:            aws.String(secretARN),
		VersionStage:        aws.String(versionstage.Current.String()),
		MoveToVersionId:     aws.String(versionID),
		RemoveFromVersionId: aws.String(versionCurrent),
	})
	if err != nil {
		return err
	}
	r.logger.Info("finishSecret: Successfully set VersionStage=%q for VersionID=%q, SecretARN=%q", versionstage.Current, versionID, secretARN)

	return nil
}

func SvcSecretManagerDefault() AWSSecretsManager {
	// Start AWS session using env vars automatically set by Lambda
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("error making AWS session: %s", err)
	}
	return secretsmanager.New(sess)
}
