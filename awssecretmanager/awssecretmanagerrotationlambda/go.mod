module github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerrotationlambda

go 1.19

replace github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib => ../awssecretmanagerlib

require (
	github.com/aws/aws-lambda-go v1.38.0
	github.com/aws/aws-sdk-go v1.44.224
	github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib v0.0.0-20230317234014-7a9c282bb368
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
