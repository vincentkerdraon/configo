module github.com/vincentkerdraon/configo/awssecretmanager

go 1.19

replace (
	github.com/vincentkerdraon/configo => ../
	github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib => ./awssecretmanagerlib
)

require (
	github.com/aws/aws-sdk-go v1.44.227
	github.com/vincentkerdraon/configo v0.3.6-0.20230322232810-3c85ac8ed431
	github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib v0.0.0-20230322232810-3c85ac8ed431
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
