package awsinstancetag_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/vincentkerdraon/configo/awsinstancetag"
)

func Example_Load() {
	//Do your own session and client init.
	//This code is expecting the instance role allows ec2 metadata access.

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	client := imds.NewFromConfig(cfg)

	iid, dio, err := awsinstancetag.Load(context.Background(), client, nil)

	if errors.Is(err, awsinstancetag.UnreachableInstanceIdentityDocumentError{}) {
		fmt.Printf("Probably not running on aws instance")
		os.Exit(1)
	}
	if errors.Is(err, awsinstancetag.ForbiddenInstanceTagReadingError{}) {
		fmt.Printf("IAM missing credentials or instance metadata not configured")
		os.Exit(1)
	}
	if err != nil {
		panic(err)
	}

	fmt.Printf("InstanceIdentityDocumentOutput: %+v\nDescribeInstancesOutput:%+v\n", iid, dio)

	/* Example output (using json)

	InstanceIdentityDocumentOutput:
	{
		"devpayProductCodes": null,
		"marketplaceProductCodes": null,
		"availabilityZone": "us-west-2d",
		"privateIp": "10.0.13.248",
		"version": "2017-09-30",
		"region": "us-west-2",
		"instanceId": "i-0847b58015a5c500f",
		"billingProducts": null,
		"instanceType": "t3a.nano",
		"accountId": "[redacted]",
		"pendingTime": "2022-10-07T20:37:15Z",
		"imageId": "ami-07eeacb3005b9beae",
		"kernelId": "",
		"ramdiskId": "",
		"architecture": "x86_64",
		"ResultMetadata": {

		}
	}

	DescribeInstancesOutput:
	{
		"NextToken": null,
		"Reservations": [
			{
			"Groups": null,
			"Instances": [
				{
				"AmiLaunchIndex": 0,
				"Architecture": "x86_64",
				"BlockDeviceMappings": [
					{
					"DeviceName": "/dev/sda1",
					"Ebs": {
						"AttachTime": "2022-10-07T20:37:15Z",
						"DeleteOnTermination": true,
						"Status": "attached",
						"VolumeId": "vol-04462e78044cf4df3"
					}
					}
				],
				"BootMode": null,
				"CapacityReservationId": null,
				"CapacityReservationSpecification": {
					"CapacityReservationPreference": "open",
					"CapacityReservationTarget": null
				},
				"ClientToken": "[redacted]",
				"CpuOptions": {
					"CoreCount": 1,
					"ThreadsPerCore": 2
				},
				"EbsOptimized": false,
				"ElasticGpuAssociations": null,
				"ElasticInferenceAcceleratorAssociations": null,
				"EnaSupport": true,
				"EnclaveOptions": {
					"Enabled": false
				},
				"HibernationOptions": {
					"Configured": false
				},
				"Hypervisor": "xen",
				"IamInstanceProfile": {
					"Arn": "arn:aws:iam::[redacted]:instance-profile/[redacted]",
					"Id": "AIPAWNLTBWAKJPM7QQ6DH"
				},
				"ImageId": "ami-07eeacb3005b9beae",
				"InstanceId": "i-0847b58015a5c500f",
				"InstanceLifecycle": null,
				"InstanceType": "t3a.nano",
				"Ipv6Address": null,
				"KernelId": null,
				"KeyName": "terraform-deploy-docker",
				"LaunchTime": "2022-10-07T20:37:15Z",
				"Licenses": null,
				"MaintenanceOptions": {
					"AutoRecovery": "default"
				},
				"MetadataOptions": {
					"HttpEndpoint": "enabled",
					"HttpProtocolIpv6": "disabled",
					"HttpPutResponseHopLimit": 1,
					"HttpTokens": "optional",
					"InstanceMetadataTags": "disabled",
					"State": "applied"
				},
				"Monitoring": {
					"State": "disabled"
				},
				"NetworkInterfaces": [
					{
					"Association": {
						"CarrierIp": null,
						"CustomerOwnedIp": null,
						"IpOwnerId": "amazon",
						"PublicDnsName": "ec2-34-222-172-161.us-west-2.compute.amazonaws.com",
						"PublicIp": "34.222.172.161"
					},
					"Attachment": {
						"AttachTime": "2022-10-07T20:37:15Z",
						"AttachmentId": "eni-attach-0004cd4d69b255df5",
						"DeleteOnTermination": true,
						"DeviceIndex": 0,
						"NetworkCardIndex": 0,
						"Status": "attached"
					},
					"Description": "",
					"Groups": [
						{
						"GroupId": "sg-04d08a116479b7e2a",
						"GroupName": "[redacted]"
						},
						{
						"GroupId": "sg-00d6e23a91228356d",
						"GroupName": "[redacted]"
						},
						{
						"GroupId": "sg-07ec9f601821d91fd",
						"GroupName": "default"
						}
					],
					"InterfaceType": "interface",
					"Ipv4Prefixes": null,
					"Ipv6Addresses": null,
					"Ipv6Prefixes": null,
					"MacAddress": "0e:81:ab:b6:f5:cf",
					"NetworkInterfaceId": "eni-0aaf145c0bfa20a5c",
					"OwnerId": "[redacted]",
					"PrivateDnsName": "ip-10-0-13-248.us-west-2.compute.internal",
					"PrivateIpAddress": "10.0.13.248",
					"PrivateIpAddresses": [
						{
						"Association": {
							"CarrierIp": null,
							"CustomerOwnedIp": null,
							"IpOwnerId": "amazon",
							"PublicDnsName": "ec2-34-222-172-161.us-west-2.compute.amazonaws.com",
							"PublicIp": "34.222.172.161"
						},
						"Primary": true,
						"PrivateDnsName": "ip-10-0-13-248.us-west-2.compute.internal",
						"PrivateIpAddress": "10.0.13.248"
						}
					],
					"SourceDestCheck": true,
					"Status": "in-use",
					"SubnetId": "subnet-0241017e705bcf4f1",
					"VpcId": "vpc-0b60b26189375f04d"
					}
				],
				"OutpostArn": null,
				"Placement": {
					"Affinity": null,
					"AvailabilityZone": "us-west-2d",
					"GroupName": "",
					"HostId": null,
					"HostResourceGroupArn": null,
					"PartitionNumber": null,
					"SpreadDomain": null,
					"Tenancy": "default"
				},
				"Platform": null,
				"PlatformDetails": "Linux/UNIX",
				"PrivateDnsName": "ip-10-0-13-248.us-west-2.compute.internal",
				"PrivateDnsNameOptions": {
					"EnableResourceNameDnsAAAARecord": false,
					"EnableResourceNameDnsARecord": false,
					"HostnameType": "ip-name"
				},
				"PrivateIpAddress": "10.0.13.248",
				"ProductCodes": null,
				"PublicDnsName": "ec2-34-222-172-161.us-west-2.compute.amazonaws.com",
				"PublicIpAddress": "34.222.172.161",
				"RamdiskId": null,
				"RootDeviceName": "/dev/sda1",
				"RootDeviceType": "ebs",
				"SecurityGroups": [
					{
					"GroupId": "sg-04d08a116479b7e2a",
					"GroupName": "[redacted]"
					},
					{
					"GroupId": "sg-00d6e23a91228356d",
					"GroupName": "[redacted]"
					},
					{
					"GroupId": "sg-07ec9f601821d91fd",
					"GroupName": "default"
					}
				],
				"SourceDestCheck": true,
				"SpotInstanceRequestId": null,
				"SriovNetSupport": null,
				"State": {
					"Code": 16,
					"Name": "running"
				},
				"StateReason": null,
				"StateTransitionReason": "",
				"SubnetId": "subnet-0241017e705bcf4f1",
				"Tags": [
					{
					"Key": "Terraform",
					"Value": "[redacted]"
					},
					{
					"Key": "EnvironmentName",
					"Value": "dev"
					},
					{
					"Key": "EnvironmentType",
					"Value": "dev"
					},
					{
					"Key": "Region",
					"Value": "us-west-2"
					},
					{
					"Key": "Name",
					"Value": "[redacted]"
					},
					{
					"Key": "Project",
					"Value": "[redacted]"
					}
				],
				"TpmSupport": null,
				"UsageOperation": "RunInstances",
				"UsageOperationUpdateTime": "2022-10-07T20:37:15Z",
				"VirtualizationType": "hvm",
				"VpcId": "vpc-0b60b26189375f04d"
				}
			],
			"OwnerId": "[redacted]",
			"RequesterId": null,
			"ReservationId": "r-0dbceff23b1f550b9"
			}
		]
		}
	*/
}
