package awssecretmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/vincentkerdraon/configo/awssecretmanager"
)

func TestIntegration_LoadRotatingSecretWhenJSON(t *testing.T) {
	t.Skip("integration test, requires credentials and secret in Secret Manager")

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			// Credentials: credentials.NewStaticCredentials("", "", ""),
			// Region: aws.String("us-east-1"),
		},
	}))
	svcSecretManager := secretsmanager.New(sess)

	sm := awssecretmanager.New(svcSecretManager)

	secretName := "prod/app"
	keyJSON := "TokenChecker.Secret"

	secret, _, err := sm.LoadRotatingSecretWhenJSON(context.Background(), secretName, keyJSON)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("SecretName=%q KeyJSON=%q Previous=%q Current=%q Pending=%q\n", secretName, keyJSON, secret.Previous, secret.Current, secret.Pending)
}
