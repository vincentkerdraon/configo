module github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerrotationlambda

go 1.23

replace github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib => ../awssecretmanagerlib

require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go v1.55.5
	github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib v0.0.0-20241015204525-2ec29f9b2420

)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
