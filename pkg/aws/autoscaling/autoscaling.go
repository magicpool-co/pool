package autoscaling

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/types"
)

func GetGroupInstanceIPs(client *aws.Client, name string) ([]string, error) {
	asgSvc := autoscaling.New(client.Session())
	res, err := asgSvc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			types.StringPtr(name),
		},
	})
	if err != nil {
		return nil, err
	} else if len(res.AutoScalingGroups) != 1 {
		return nil, fmt.Errorf("group size mismatch: have %d, want 1", len(res.AutoScalingGroups))
	}

	group := res.AutoScalingGroups[0]
	ids := make([]*string, len(group.Instances))
	for i, instance := range group.Instances {
		ids[i] = instance.InstanceId
	}

	ec2Svc := ec2.New(client.Session())
	instances, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: ids,
	})
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0)
	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			ips = append(ips, types.StringValue(instance.PrivateIpAddress))
		}
	}

	if len(ips) != len(ids) {
		return nil, fmt.Errorf("instance size mismatch: have %d, want %d", len(ips), len(ids))
	}

	return ips, nil
}
