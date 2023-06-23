package awssecretmanager_test

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/vincentkerdraon/configo/awssecretmanager"
	"github.com/vincentkerdraon/configo/config"
	"github.com/vincentkerdraon/configo/config/param"
	"github.com/vincentkerdraon/configo/lock"
	"github.com/vincentkerdraon/configo/secretrotation"
)

type applicationConfiguration struct {
	DynamoTable string
	APISecret   secretrotation.Manager
}

func initConfigWhenLoadValueWhenPlainText(svcSecretManager awssecretmanager.AWSSecretsManager) func() *applicationConfiguration {
	appConfig := applicationConfiguration{}
	sm := awssecretmanager.New(svcSecretManager)

	//Simple read from the secret manager (plain text + not rotating)

	pDynamoTable, err := param.New("DynamoTable",
		func(s string) error { appConfig.DynamoTable = s; return nil },
		param.WithLoader(
			func(ctx context.Context) (string, error) {
				s, _, err := sm.LoadValueWhenPlainText(ctx, "unit_test_rotation_secret_manager_DynamoTable")
				if err != nil {
					return "", err
				}
				return s.String(), nil
			},
			//Check regularly in addition to startup
			// (For this example: super fast)
			param.WithSynchroFrequency(100*time.Millisecond),
		),
	)
	if err != nil {
		panic(err)
	}

	lock := lock.New()
	c, err := config.New(
		config.WithParams(pDynamoTable),
		config.WithLock(lock),
	)
	if err != nil {
		panic(err)
	}
	err = c.Init(
		context.Background(),
		//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
		config.WithInputArgs([]string{}),
	)
	if err != nil {
		panic(err)
	}

	return func() *applicationConfiguration {
		lock.Lock()
		defer lock.Unlock()
		return &appConfig
	}
}

func initConfigWhenLoadRotatingSecretWhenJSON(svcSecretManager awssecretmanager.AWSSecretsManager) func() *applicationConfiguration {
	appConfig := applicationConfiguration{}
	sm := awssecretmanager.New(svcSecretManager)

	// Read a rotating secret from the secret manager. (JSON)
	// For this example: The name of the secret to use comes from the PreConfig.

	pAPISecret, err := param.New("APISecret",
		func(s string) error {
			rs := &secretrotation.RotatingSecret{}
			if err := rs.Deserialize(s); err != nil {
				return err
			}
			return appConfig.APISecret.Set(*rs)
		},
		param.WithLoader(
			func(ctx context.Context) (string, error) {
				secret, _, err := sm.LoadRotatingSecretWhenJSON(ctx, "unit_test_rotation_secret_manager_APISecret", "SecretID")
				if err != nil {
					return "", err
				}
				return secret.Serialize(), err
			},
			//Check regularly in addition to startup
			param.WithSynchroFrequency(time.Minute),
		),
	)
	if err != nil {
		panic(err)
	}

	lock := lock.New()
	c, err := config.New(
		config.WithParams(pAPISecret),
		config.WithLock(lock),
	)
	if err != nil {
		panic(err)
	}
	err = c.Init(
		context.Background(),
		//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
		config.WithInputArgs([]string{}),
	)
	if err != nil {
		panic(err)
	}

	return func() *applicationConfiguration {
		lock.Lock()
		defer lock.Unlock()
		return &appConfig
	}
}

func initAws() *secretsmanager.SecretsManager {
	//To execute this test, assuming you have AWS credentials. Set your own credential configuration.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			// Credentials: credentials.New NewStaticCredentials("AKIAYTQSIPFQQ4OJSYW3", "", ""),
			Region: aws.String("us-east-1"),
		},
	}))
	svcSecretManager := secretsmanager.New(sess)
	//Making sure the secret exists (and basic check on the permissions)
	_, err := svcSecretManager.DescribeSecret(&secretsmanager.DescribeSecretInput{SecretId: aws.String("unit_test_rotation_secret_manager_APISecret")})
	if err != nil {
		panic(err)
	}

	return svcSecretManager
}

func Exampleimpl_LoadValueWhenPlainText() {
	svcSecretManager := initAws()

	//Read a smaller config first to find the params required in the larger config
	configProxy := initConfigWhenLoadValueWhenPlainText(svcSecretManager)

	//using a function with the lock to avoid race condition on value read.
	//This is only needed when using Loader with Synchro.
	//See also lock.LockWithContext(ctx)

	fmt.Printf("DynamoTable=%q\n", configProxy().DynamoTable)

	// // Output:
	// // DynamoTable="myTableName"
}

func Exampleimpl_LoadRotatingSecretWhenJSON() {
	svcSecretManager := initAws()

	//Read a smaller config first to find the params required in the larger config
	configProxy := initConfigWhenLoadRotatingSecretWhenJSON(svcSecretManager)

	//using a function with the lock to avoid race condition on value read.
	//This is only needed when using Loader with Synchro.
	//See also lock.LockWithContext(ctx)

	fmt.Printf("%+v\n", configProxy())
	secret, err := configProxy().APISecret.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("APISecret.Current: %s\n", secret)

	// // Output:
	// // {Name:Vincent City:Vancouver Age:35 nonExportedField:}
}
