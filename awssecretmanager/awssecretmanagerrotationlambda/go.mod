module github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerrotationlambda

go 1.21

replace github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib => ../awssecretmanagerlib

require (
	github.com/aws/aws-lambda-go v1.44.0
	github.com/aws/aws-sdk-go v1.49.17
	github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib v0.0.0-20231120164007-846e044da8e1
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
