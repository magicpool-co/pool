package ecs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/types"
)

func GetContainerByInstanceID(client *aws.Client, cluster, instanceID string) (string, error) {
	ecsSvc := ecs.New(client.Session())
	containers, err := ecsSvc.ListContainerInstances(&ecs.ListContainerInstancesInput{
		Cluster: types.StringPtr(cluster),
		Filter:  types.StringPtr(fmt.Sprintf(`ec2InstanceId == '%s'`, instanceID)),
	})
	if err != nil {
		return "", err
	} else if len(containers.ContainerInstanceArns) == 0 {
		return "", fmt.Errorf("unable to find container by instance %s in cluster %s", instanceID, cluster)
	}

	return types.StringValue(containers.ContainerInstanceArns[0]), nil
}

func toggleDrainInstance(client *aws.Client, cluster, instance string, drain bool) error {
	state := "ACTIVE"
	if drain {
		state = "DRAINING"
	}

	ecsSvc := ecs.New(client.Session())
	res, err := ecsSvc.UpdateContainerInstancesState(&ecs.UpdateContainerInstancesStateInput{
		Cluster: types.StringPtr(cluster),
		ContainerInstances: []*string{
			types.StringPtr(instance),
		},
		Status: types.StringPtr(state),
	})
	if err != nil {
		return err
	}

	for _, failure := range res.Failures {
		arn := types.StringValue(failure.Arn)
		detail := types.StringValue(failure.Detail)
		reason := types.StringValue(failure.Reason)

		return fmt.Errorf("failed for ARN (%s): %s: %s", arn, detail, reason)
	}

	return nil
}
func DrainClusterContainerInstance(client *aws.Client, cluster, instance string) error {
	return toggleDrainInstance(client, cluster, instance, true)
}

func ActivateClusterContainerInstance(client *aws.Client, cluster, instance string) error {
	return toggleDrainInstance(client, cluster, instance, false)
}
