module github.com/vincentkerdraon/configo/awssecretmanager

go 1.21

replace (
	github.com/vincentkerdraon/configo => ../
	github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib => ./awssecretmanagerlib
)

require (
	github.com/aws/aws-sdk-go v1.48.0
	github.com/vincentkerdraon/configo v0.0.0-00010101000000-000000000000
	github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib v0.0.0-00010101000000-000000000000
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
