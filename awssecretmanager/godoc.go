/*
Package awssecretmanager helps loading a secret from https://aws.amazon.com/secrets-manager/

Helper for the default format available from the console:
  - plain text
  - JSON.

Rotation state:

  - disable: there is only one value.

  - enable: a lambda is rotating the secret. Retriving values for Previous + Current + Pending

    ----

Policy setup + lambda setup

https://github.com/aws-samples/aws-secrets-manager-rotation-lambdas
https://github.com/aws-samples/aws-secrets-manager-rotation-lambdas/blob/master/SecretsManagerRotationTemplate/lambda_function.py
https://github.com/square/password-rotation-lambda/blob/v1.0.1/examples/rds/main.go
https://pkg.go.dev/github.com/square/password-rotation-lambda@v1.0.1
*/
package awssecretmanager
