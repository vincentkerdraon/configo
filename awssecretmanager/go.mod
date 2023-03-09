module github.com/vincentkerdraon/configo/awssecretmanager

go 1.19

replace github.com/vincentkerdraon/configo => ../

require (
	github.com/aws/aws-sdk-go v1.44.216
	github.com/vincentkerdraon/configo v0.0.0-20230125233039-75cbc84c4bb3
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
