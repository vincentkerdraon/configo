// Package awsinstancetag helps retrieving data from AWS instance metadata.
//
// This is oriented toward software configuration (not infrastructure checking or monitoring).
// Secrets must NOT be stored in metadata (not safe).
// Some methods require tags to be explicitly allowed in the instance options.
//
// See https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html and https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html
//
// This is opinionated for a specific use:
//   - Using https://ec2.amazonaws.com/?Action=DescribeInstances (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html)
//   - Using "http://169.254.169.254/latest/dynamic/instance-identity/document"
//   - NOT using "http://169.254.169.254/latest/user-data"
//   - NOT using "http://169.254.169.254/latest/meta-data"
//
// It returns a custom error for common catch errors (Like "not on AWS" or "instance metadata not configured")
package awsinstancetag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	AWSInstanceMetadataService interface {
		GetInstanceIdentityDocument(ctx context.Context, params *imds.GetInstanceIdentityDocumentInput, optFns ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error)
	}

	AWSEC2Service interface {
		DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	}
)

// Load gets instances metadata.
//
// First, it gets the InstanceIdentityDocument for the InstanceID + region. Then it creates a new session with the region and finally it calls DescribeInstances.
//
// It uses CreateEC2sDefault() if ec2s is null.
func Load(
	ctx context.Context,
	ims AWSInstanceMetadataService,
	ec2s func(region string) (AWSEC2Service, error),
) (
	*imds.GetInstanceIdentityDocumentOutput,
	*ec2.DescribeInstancesOutput,
	error,
) {
	//find instance ID + region (needed to call ec2 DescribeInstances)
	idDoc, err := ims.GetInstanceIdentityDocument(ctx, nil)
	//There is no real way to differentiate the context deadline exceeded.
	//I am assuming that if it does not answer, then it is not reachable.
	//This is the error we receive:
	//%T: *smithy.OperationError
	//%s: operation error ec2imds: GetInstanceIdentityDocument, request canceled, context deadline exceeded
	//%#v: &smithy.OperationError{ServiceID:"ec2imds", OperationName:"GetInstanceIdentityDocument", Err:(*aws.RequestCanceledError)(0xc000024250)}
	if errors.Is(err, context.DeadlineExceeded) {
		return nil, nil, UnreachableInstanceIdentityDocumentError{Err: err}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("fail retrieve InstanceIdentityDocument from the EC2 instance: %w", err)
	}

	//Creating another session to call ec2 DescribeInstances
	//The session creation is injected, because there are many ways of doing it.
	//Default assumes the credentials are already in the chain.
	if ec2s == nil {
		ec2s = CreateEC2sDefault
	}

	svc, err := ec2s(idDoc.Region)
	if err != nil {
		return nil, nil, err
	}
	instanceInfo, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(idDoc.InstanceID)},
	})
	//This error happens if the tags access is not configured on the instance.
	//%T: *awserr.requestError
	//%s: UnauthorizedOperation: You are not authorized to perform this operation.	status code: 403, request id: f2bb4e20-3720-42e1-ba67-3f952b98e5ec
	//%#v: &awserr.requestError{awsError:(*awserr.baseError)(0xc0000ec840), statusCode:403, requestID:"f2bb4e20-3720-42e1-ba67-3f952b98e5ec", bytes:[]uint8(nil)}
	if err != nil && strings.Contains(err.Error(), "UnauthorizedOperation") {
		return nil, nil, ForbiddenInstanceTagReadingError{Err: err}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("fail DescribeInstances InstanceID:%s, Region:%s, %v", idDoc.InstanceID, idDoc.Region, err)
	}

	return idDoc, instanceInfo, nil
}

func CreateEC2sDefault(region string) (AWSEC2Service, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("fail create NewSession in %q: %w", region, err)
	}
	return ec2.New(sess), nil
}
