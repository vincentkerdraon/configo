package versionstage

// VersionStage is the different version for a same secret (the one before, the current one, the future one)
//
// from github.com/aws/aws-sdk-go@v1.44.216/service/secretsmanager/api.go
// because stupid AWS lib won't export it
type VersionStage string

const (
	Previous VersionStage = "AWSPREVIOUS"
	Current  VersionStage = "AWSCURRENT"
	Pending  VersionStage = "AWSPENDING"
)

func (s VersionStage) String() string {
	return string(s)
}
