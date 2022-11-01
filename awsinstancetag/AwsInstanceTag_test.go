package awsinstancetag

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type imsMock struct{}

func (ims imsMock) GetInstanceIdentityDocument(ctx context.Context, params *imds.GetInstanceIdentityDocumentInput, optFns ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error) {
	return &imds.GetInstanceIdentityDocumentOutput{
		InstanceIdentityDocument: imds.InstanceIdentityDocument{
			Region:     "reg",
			InstanceID: "iid",
		},
	}, nil
}

type ec2sMock struct{}

func (ec2s ec2sMock) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	tag1 := "Tag1"
	val1 := "Val1"
	return &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					{
						Tags: []*ec2.Tag{
							{
								Key:   &tag1,
								Value: &val1,
							},
						},
					},
				},
			},
		},
	}, nil
}

func TestLoad(t *testing.T) {
	ec2s := &ec2sMock{}
	ims := &imsMock{}

	type args struct {
		ims  AWSInstanceMetadataService
		ec2s func(region string) (AWSEC2Service, error)
	}
	tests := []struct {
		name  string
		args  args
		check func(*testing.T, *imds.GetInstanceIdentityDocumentOutput, *ec2.DescribeInstancesOutput, error)
	}{
		{
			name: "ok",
			args: args{
				ims:  ims,
				ec2s: func(region string) (AWSEC2Service, error) { return ec2s, nil },
			},
			check: func(t *testing.T, giido *imds.GetInstanceIdentityDocumentOutput, dio *ec2.DescribeInstancesOutput, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if giido.Region != "reg" || giido.InstanceID != "iid" {
					t.Error()
				}
				if *dio.Reservations[0].Instances[0].Tags[0].Key != "Tag1" {
					t.Error()
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIID, gotDIO, err := Load(context.Background(), tt.args.ims, tt.args.ec2s)
			tt.check(t, gotIID, gotDIO, err)
		})
	}
}
