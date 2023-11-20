module github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerrotationlambda

go 1.21

replace github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib => ../awssecretmanagerlib

require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/aws/aws-sdk-go v1.48.0
	github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib v0.0.0-00010101000000-000000000000
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
