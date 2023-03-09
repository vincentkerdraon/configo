/*
Package awssecretmanagerrotationlambda rotates a secret in https://aws.amazon.com/secrets-manager/

Following https://docs.aws.amazon.com/secretsmanager/latest/userguide/rotating-secrets.html

And mostly based on
- https://github.com/aws-samples/aws-secrets-manager-rotation-lambdas/blob/master/SecretsManagerRotationTemplate/lambda_function.py
- https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html

Steps to use:
  - create secret manager entry with initial values
  - deploy lambda https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html
  - GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go
  - zip main.zip main
  - deploy lambda. Maybe with env var if you have a complex implementation.
  - give permissions for the secret manager to use the lambda
  - configure secret manager entry to call lambda
  - test by requesting a rotation, validate the rotated value.
*/
package awssecretmanagerrotationlambda
