package awssecretmanagerrotationlambda_test

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/lambdaconf"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerrotationlambda"
)

func Example_simple() {
	//No argument in New(): using all the defaults
	rotater := awssecretmanagerrotationlambda.New()
	lambda.Start(rotater.HandleRequest)
}

func Example_allOptions() {
	//Do you own logic to change the secret old value into the new value
	//Useful if you are using a JSON document instead of only one value
	prepareSecretNew := func(ctx context.Context, secretARN, secretOld string) (secretNew string, _ error) {
		//(this is stupid simple logic for the example)
		return secretOld + "-", nil
	}

	//for example if you want the lambda to change the database secret (directly in the database service).
	setSecret := func(ctx context.Context, secretARN, versionID string) error {
		return nil
	}
	//if you used setSecret(), you want to test if it worked.
	testSecret := func(ctx context.Context, secretARN, versionID string) error {
		return nil
	}

	rotater := awssecretmanagerrotationlambda.New(
		awssecretmanagerrotationlambda.WithPrepareSecret(prepareSecretNew),
		awssecretmanagerrotationlambda.WithSetSecret(setSecret),
		awssecretmanagerrotationlambda.WithTestSecret(testSecret),

		// You could also inject AWSSecretsManagerService if needed
		// awssecretmanagerrotationlambda.WithAWSSecretsManager(svc),
	)
	lambda.Start(rotater.HandleRequest)
}

func Example_withComplexJSONValues() {
	//let's say you have a secret manager entry:
	// - arn = SecretID1
	// - content = {"key1":"val1","key2":"val2"}
	// we want the lambda to rotate "val1" but leave everything else untouched

	//ignore this. Assume configuration is set for the lambda
	err := os.Setenv("Conf", `{"Secrets":{"secretARN1":{"Keys":{"key1":{"Constraint":"AlphaNum","AlphaNumLength":16}}}}}`)
	if err != nil {
		panic(err)
	}

	//read lambda configuration
	lambdaConf := lambdaconf.LambdaConf{}
	lambdaConfJSON := os.Getenv("Conf")
	if err := json.Unmarshal([]byte(lambdaConfJSON), &lambdaConf); err != nil {
		panic(err)
	}
	if err := lambdaConf.Validate(); err != nil {
		panic(err)
	}

	//define the function to change the secret in a JSON document.
	prepareSecretNew := lambdaconf.PrepareNewSecretFormatted(time.Now(), lambdaConf)

	//Create and start the lambda listener
	rotater := awssecretmanagerrotationlambda.New(awssecretmanagerrotationlambda.WithPrepareSecret(prepareSecretNew))
	lambda.Start(rotater.HandleRequest)

	// after running the lambda, the secret will change to something similar to:
	// - content = {"key1":"vqYXDhvE0oG0Smbj","key2":"val2"}
}
