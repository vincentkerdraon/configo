module github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerrotationlambda

go 1.19

replace github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib => ../awssecretmanagerlib

require (
	github.com/aws/aws-lambda-go v1.38.0
	github.com/aws/aws-sdk-go v1.44.216
	github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib v0.0.0-00010101000000-000000000000
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
