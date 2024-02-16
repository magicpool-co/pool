package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/types"
)

func GetInstanceIPByID(client *aws.Client, id string) (string, error) {
	ec2Svc := ec2.New(client.Session())
	instances, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{types.StringPtr(id)},
	})
	if err != nil {
		return "", err
	}

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			return types.StringValue(instance.PrivateIpAddress), nil
		}
	}

	return "", fmt.Errorf("unable to find instance from id %s", id)
}

func GetInstanceIDByIP(client *aws.Client, ip string) (string, error) {
	ec2Svc := ec2.New(client.Session())
	instances, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: types.StringPtr("private-ip-address"),
				Values: []*string{
					types.StringPtr(ip),
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			return types.StringValue(instance.InstanceId), nil
		}
	}

	return "", fmt.Errorf("unable to find instance from ip %s", ip)
}
