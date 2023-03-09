package versionstage

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
