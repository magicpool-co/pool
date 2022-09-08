package ec2

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/types"
)

func SendCommandToInstance(client *aws.Client, instance string, cmds []string) (string, error) {
	if len(cmds) == 0 {
		return "", nil
	}

	cmdPtrs := make([]*string, len(cmds))
	for i, cmd := range cmds {
		cmdPtrs[i] = types.StringPtr(cmd)
	}

	ssmSvc := ssm.New(client.Session())
	res, err := ssmSvc.SendCommand(&ssm.SendCommandInput{
		DocumentName: types.StringPtr("AWS-RunShellScript"),
		InstanceIds: []*string{
			types.StringPtr(instance),
		},
		Parameters: map[string][]*string{
			"commands": cmdPtrs,
		},
		TimeoutSeconds: types.Int64Ptr(60 * 60 * 4),
	})
	if err != nil {
		return "", err
	}

	return types.StringValue(res.Command.CommandId), nil
}

func WaitForCommand(client *aws.Client, instance, commandID string) (string, error) {
	if commandID == "" {
		return "", nil
	}

	ssmSvc := ssm.New(client.Session())
	for {
		time.Sleep(time.Second)
		res, err := ssmSvc.GetCommandInvocation(&ssm.GetCommandInvocationInput{
			CommandId:  types.StringPtr(commandID),
			InstanceId: types.StringPtr(instance),
		})
		if err != nil {
			return "", err
		}

		status := types.StringValue(res.Status)
		switch status {
		case "Pending", "InProgress", "Delayed":
		case "TimedOut", "Failed", "Cancelled", "Cancelling":
			msg := status
			if reason := types.StringValue(res.StandardErrorContent); len(reason) > 0 {
				msg += " :" + reason
			}
			if url := types.StringValue(res.StandardErrorUrl); len(url) > 0 {
				msg += " :" + url
			}
			if code := types.Int64Value(res.ResponseCode); code != 0 {
				msg += " (" + strconv.Itoa(int(code)) + ")"
			}

			return "", fmt.Errorf("unable to complete command: %s", msg)
		case "Success":
			return types.StringValue(res.StandardOutputContent), nil
		default:
			return "", fmt.Errorf("unknown command status %s", status)
		}
	}
}
