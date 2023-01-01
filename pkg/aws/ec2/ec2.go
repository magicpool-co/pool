package ec2

import (
	"fmt"
	"strings"
	"time"

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

func GetInstanceVolumes(client *aws.Client, instance string) ([]string, error) {
	ec2Svc := ec2.New(client.Session())
	instances, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{types.StringPtr(instance)},
	})
	if err != nil {
		return nil, err
	}

	volumeIDs := make([]string, 0)
	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			for _, device := range instance.BlockDeviceMappings {
				volumeIDs = append(volumeIDs, types.StringValue(device.Ebs.VolumeId))
			}
		}
	}

	return volumeIDs, nil
}

func GetEBSVolumeSize(client *aws.Client, volumeID string) (int64, error) {
	ec2Svc := ec2.New(client.Session())
	volumes, err := ec2Svc.DescribeVolumes(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{types.StringPtr(volumeID)},
	})
	if err != nil {
		return 0, err
	} else if len(volumes.Volumes) == 0 {
		return 0, fmt.Errorf("unable to find volume %s", volumeID)
	}

	return types.Int64Value(volumes.Volumes[0].Size), nil
}

func ResizeInstanceVolume(client *aws.Client, instance string, size int64) error {
	volumeIDs, err := GetInstanceVolumes(client, instance)
	if err != nil {
		return err
	} else if len(volumeIDs) != 1 {
		return fmt.Errorf("unable to find volume")
	}

	// check current volume size
	currentSize, err := GetEBSVolumeSize(client, volumeIDs[0])
	if err != nil {
		return err
	} else if size <= currentSize {
		return fmt.Errorf("volume size must be greater than current size")
	}

	// check current modifications
	ec2Svc := ec2.New(client.Session())
	modRes, err := ec2Svc.DescribeVolumesModifications(&ec2.DescribeVolumesModificationsInput{
		VolumeIds: []*string{types.StringPtr(volumeIDs[0])},
	})
	if err != nil {
		return err
	}

	for _, mod := range modRes.VolumesModifications {
		switch strings.ToLower(types.StringValue(mod.ModificationState)) {
		case "modifying", "optimizing":
			return fmt.Errorf("cannot modify volume - already being modified")
		}
	}

	res, err := ec2Svc.ModifyVolume(&ec2.ModifyVolumeInput{
		Size:     types.Int64Ptr(size),
		VolumeId: types.StringPtr(volumeIDs[0]),
	})
	if err != nil {
		return err
	}

	resizeCmds := []string{"growpart /dev/nvme0n1 1", "resize2fs /dev/nvme0n1p1"}
	startTime := types.TimeValue(res.VolumeModification.StartTime)
	if startTime.IsZero() {
		return fmt.Errorf("nil start time on resizing instance %s", instance)
	}

	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   types.StringPtr("start-time"),
			Values: []*string{types.StringPtr(startTime.String())},
		},
	}

	for {
		time.Sleep(time.Second * 1)
		res, err := ec2Svc.DescribeVolumesModifications(&ec2.DescribeVolumesModificationsInput{
			VolumeIds: []*string{types.StringPtr(volumeIDs[0])},
			Filters:   filters,
		})
		if err != nil {
			return err
		} else if len(res.VolumesModifications) == 0 {
			continue
		}

		state := strings.ToLower(types.StringValue(res.VolumesModifications[0].ModificationState))
		switch state {
		case "modifying":
		case "completed", "optimizing":
			commandID, err := SendCommandToInstance(client, instance, resizeCmds)
			if err != nil {
				return err
			} else if _, err := WaitForCommand(client, instance, commandID); err != nil {
				return err
			}
			return nil
		case "failed":
			return fmt.Errorf("modification of volume %s on instance %s failed", volumeIDs[0], instance)
		}
	}
}
