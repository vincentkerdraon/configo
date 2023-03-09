/*
Package awssecretmanagerrotationlambda rotates a secret in https://aws.amazon.com/secrets-manager/

Following https://docs.aws.amazon.com/secretsmanager/latest/userguide/rotating-secrets.html and adding options to deal with JSON entries

And mostly based on:
  - https://github.com/aws-samples/aws-secrets-manager-rotation-lambdas/blob/master/SecretsManagerRotationTemplate/lambda_function.py
  - https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html

Steps to use:
  - create secret manager entry with initial values
  - deploy lambda https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html
  - => GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go
  - => zip main.zip main
  - deploy lambda. Maybe with env var if you have a complex implementation.
  - => upload .zip file
  - => HandlerInfo is the name of the binary file to use. `main` in this case
  - => use minimum perf lambda, Memory 128MB
  - => set Permissions.Execution role
  - => set Permissions.Resource-based policy statements
  - configure secret manager entry to call lambda in "Rotation configuration"
  - test by requesting a rotation: "Rotate secret immediately", validate the rotated value.

Permissions.Execution role: Policy SecretManagerRotateLambda. Example allowing access to all secrets.

	{
	    "Version": "2012-10-17",
	    "Statement": [
	        {
	            "Sid": "VisualEditor0",
	            "Effect": "Allow",
	            "Action": "logs:CreateLogGroup",
	            "Resource": "arn:aws:logs:us-east-1:_redacted_:*"
	        },
	        {
	            "Sid": "VisualEditor1",
	            "Effect": "Allow",
	            "Action": [
	                "logs:CreateLogStream",
	                "logs:PutLogEvents"
	            ],
	            "Resource": "arn:aws:logs:us-east-1:_redacted_:log-group:/aws/lambda/SecretManagerRotate:*"
	        },
	        {
	            "Sid": "VisualEditor2",
	            "Effect": "Allow",
	            "Action": [
	                "secretsmanager:GetRandomPassword",
	                "secretsmanager:GetSecretValue",
	                "secretsmanager:DescribeSecret",
	                "secretsmanager:PutSecretValue",
	                "secretsmanager:UpdateSecretVersionStage"
	            ],
	            "Resource": "*"
	        }
	    ]
	}

# Permissions.Resource-based policy statements

  - Statement ID:SecretManagerRotate
  - Principal: secretsmanager.amazonaws.com
  - PrincipalOrgID: -
  - Conditions: None
  - Action: lambda:InvokeFunction
*/
package awssecretmanagerrotationlambda
