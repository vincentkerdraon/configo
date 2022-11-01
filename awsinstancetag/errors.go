package awsinstancetag

import "fmt"

type UnreachableInstanceIdentityDocumentError struct {
	Err error
}

func (e UnreachableInstanceIdentityDocumentError) Error() string {
	return fmt.Sprintf("unreachable Instance Identity DocumentError, %s", e.Err)
}

func (e UnreachableInstanceIdentityDocumentError) Unwrap() error { return e.Err }

// see https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-options.html
type ForbiddenInstanceTagReadingError struct {
	Err error
}

func (e ForbiddenInstanceTagReadingError) Error() string {
	return fmt.Sprintf("need access to instance Metadata, check IAM + instance options. %s", e.Err)
}

func (e ForbiddenInstanceTagReadingError) Unwrap() error { return e.Err }
