package ecs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/types"
)

func FetchContainerIPs(client *aws.Client, cluster string) ([]string, error) {
	ips := []string{}

	ecsSvc := ecs.New(client.Session())
	containerArns, err := ecsSvc.ListContainerInstances(&ecs.ListContainerInstancesInput{
		Cluster: types.StringPtr(cluster),
	})
	if err != nil {
		return ips, err
	}

	containers, err := ecsSvc.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            types.StringPtr(cluster),
		ContainerInstances: containerArns.ContainerInstanceArns,
	})
	if err != nil {
		return ips, err
	}

	instanceIds := []*string{}
	for _, container := range containers.ContainerInstances {
		instanceIds = append(instanceIds, container.Ec2InstanceId)
	}

	ec2Svc := ec2.New(client.Session())
	instances, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return ips, err
	}

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			ips = append(ips, types.StringValue(instance.PrivateIpAddress))
		}
	}

	return ips, nil
}

func handleFailures(failures []*ecs.Failure) []error {
	errs := make([]error, len(failures))
	for i, failure := range failures {
		arn := types.StringValue(failure.Arn)
		detail := types.StringValue(failure.Detail)
		reason := types.StringValue(failure.Reason)
		errs[i] = fmt.Errorf("failed for ARN (%s): %s: %s", arn, detail, reason)
	}

	return errs
}

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

func GetTasksByContainer(client *aws.Client, cluster, container string) ([]string, error) {
	ecsSvc := ecs.New(client.Session())

	var nextToken *string
	taskArns := make([]string, 0)
	for {
		tasks, err := ecsSvc.ListTasks(&ecs.ListTasksInput{
			Cluster:           types.StringPtr(cluster),
			ContainerInstance: types.StringPtr(container),
			NextToken:         nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, arn := range tasks.TaskArns {
			taskArns = append(taskArns, types.StringValue(arn))
		}

		if len(types.StringValue(tasks.NextToken)) == 0 {
			break
		}
		nextToken = tasks.NextToken
	}

	return taskArns, nil
}

func GetServiceRevision(client *aws.Client, cluster, service string) (int64, error) {
	ecsSvc := ecs.New(client.Session())
	services, err := ecsSvc.DescribeServices(&ecs.DescribeServicesInput{
		Cluster: types.StringPtr(cluster),
		Services: []*string{
			types.StringPtr(service),
		},
	})
	if err != nil {
		return 0, err
	} else if failures := handleFailures(services.Failures); len(failures) > 0 {
		return 0, failures[0]
	} else if len(services.Services) == 0 {
		return 0, fmt.Errorf("unable to find service %s in cluster %s", service, cluster)
	}

	parts := strings.Split(types.StringValue(services.Services[0].TaskDefinition), ":")
	revision, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)

	return revision, err
}

func GetTaskRevision(client *aws.Client, cluster, taskARN string) (int64, error) {
	ecsSvc := ecs.New(client.Session())
	tasks, err := ecsSvc.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: types.StringPtr(cluster),
		Tasks: []*string{
			types.StringPtr(taskARN),
		},
	})
	if err != nil {
		return 0, err
	} else if failures := handleFailures(tasks.Failures); len(failures) > 0 {
		return 0, failures[0]
	} else if len(tasks.Tasks) == 0 {
		return 0, fmt.Errorf("unable to find task %s in cluster %s", taskARN, cluster)
	}

	return types.Int64Value(tasks.Tasks[0].Version), nil
}

func ListClusterTasksByIP(client *aws.Client, cluster string) (map[string]string, error) {
	ecsSvc := ecs.New(client.Session())

	var nextToken *string
	taskArns := make([]*string, 0)
	for {
		tasks, err := ecsSvc.ListTasks(&ecs.ListTasksInput{
			Cluster:   types.StringPtr(cluster),
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, arn := range tasks.TaskArns {
			if len(types.StringValue(arn)) > 0 {
				taskArns = append(taskArns, arn)
			}
		}

		if len(types.StringValue(tasks.NextToken)) == 0 {
			break
		}
		nextToken = tasks.NextToken
	}

	if len(taskArns) > 100 {
		return nil, fmt.Errorf("too many tasks to query")
	} else if len(taskArns) == 0 {
		return nil, nil
	}

	tasks, err := ecsSvc.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: types.StringPtr(cluster),
		Tasks:   taskArns,
	})
	if err != nil {
		return nil, err
	} else if failures := handleFailures(tasks.Failures); len(failures) > 0 {
		return nil, failures[0]
	}

	containerArns := make([]*string, 0)
	containerArnIndex := make(map[string]string, 0)
	for _, task := range tasks.Tasks {
		containerArns = append(containerArns, task.ContainerInstanceArn)
		containerArnIndex[types.StringValue(task.TaskArn)] = types.StringValue(task.ContainerInstanceArn)
	}

	if len(containerArns) > 100 {
		return nil, fmt.Errorf("too many containers to query")
	} else if len(containerArns) == 0 {
		return nil, nil
	}

	containers, err := DescribeContainerInstances(client, cluster, containerArns)
	if err != nil {
		return nil, err
	}

	instanceIds := make([]*string, len(containers))
	instanceIdIndex := make(map[string]string, len(containers))
	for i, container := range containers {
		instanceIds[i] = container.Ec2InstanceId
		instanceIdIndex[types.StringValue(container.ContainerInstanceArn)] = types.StringValue(container.Ec2InstanceId)
	}

	nextToken = nil
	instanceIpIndex := make(map[string]string)
	ec2Svc := ec2.New(client.Session())
	for {
		instances, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: instanceIds,
			NextToken:   nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, reservation := range instances.Reservations {
			for _, instance := range reservation.Instances {
				instanceIpIndex[types.StringValue(instance.InstanceId)] = types.StringValue(instance.PrivateIpAddress)
			}
		}

		if len(types.StringValue(instances.NextToken)) == 0 {
			break
		}
		nextToken = instances.NextToken
	}

	taskIpIndex := make(map[string]string)
	for task, container := range containerArnIndex {
		instance := instanceIdIndex[container]
		if len(instance) == 0 {
			continue
		}

		ip := instanceIpIndex[instance]
		if len(ip) == 0 {
			continue
		}

		taskIpIndex[ip] = task
	}

	return taskIpIndex, nil
}

func StopTask(client *aws.Client, cluster, task string) error {
	ecsSvc := ecs.New(client.Session())
	_, err := ecsSvc.StopTask(&ecs.StopTaskInput{
		Cluster: types.StringPtr(cluster),
		Task:    types.StringPtr(task),
	})

	return err
}

func DescribeContainerInstances(client *aws.Client, cluster string, instances []*string) ([]*ecs.ContainerInstance, error) {
	ecsSvc := ecs.New(client.Session())
	res, err := ecsSvc.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            types.StringPtr(cluster),
		ContainerInstances: instances,
	})
	if err != nil {
		return nil, err
	} else if failures := handleFailures(res.Failures); len(failures) > 0 {
		return nil, failures[0]
	}

	return res.ContainerInstances, nil
}

func DrainClusterContainerInstance(client *aws.Client, cluster, instance string) error {
	return toggleDrainInstance(client, cluster, instance, true)
}

func ActivateClusterContainerInstance(client *aws.Client, cluster, instance string) error {
	return toggleDrainInstance(client, cluster, instance, false)
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
	} else if failures := handleFailures(res.Failures); len(failures) > 0 {
		return failures[0]
	}

	return nil
}
