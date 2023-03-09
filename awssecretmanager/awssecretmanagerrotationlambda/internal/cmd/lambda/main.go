package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/lambdaconf"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerrotationlambda"
)

func main() {
	//read lambda configuration
	lambdaConf := lambdaconf.LambdaConf{}
	lambdaConfJSON := os.Getenv("Conf")
	if lambdaConfJSON == "" {
		panic(`require env var "Conf" with JSON LambdaConf`)
	}
	if err := json.Unmarshal([]byte(lambdaConfJSON), &lambdaConf); err != nil {
		panic(err)
	}
	if err := lambdaConf.Validate(); err != nil {
		panic(err)
	}

	//define the function to change the secret in a JSON document.
	prepareSecretNew := lambdaconf.PrepareNewSecretFormatted(time.Now(), lambdaConf)

	//Create and start the lambda listener
	rotater := awssecretmanagerrotationlambda.New(
		awssecretmanagerrotationlambda.WithLogger(awssecretmanagerrotationlambda.NewLeveledLoggerStandard(awssecretmanagerrotationlambda.LogLevelTrace)),
		awssecretmanagerrotationlambda.WithPrepareSecret(prepareSecretNew),
	)
	lambda.Start(rotater.HandleRequest)
}
