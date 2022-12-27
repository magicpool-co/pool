package sqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/types"
)

type Message struct {
	ID         string
	Attributes map[string]string
}

func PopFromQueue(client *aws.Client, queue string) ([]*Message, error) {
	svc := sqs.New(client.Session())
	url, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: types.StringPtr(queue),
	})
	if err != nil {
		return nil, err
	}

	results, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			types.StringPtr(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			types.StringPtr(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            url.QueueUrl,
		MaxNumberOfMessages: types.Int64Ptr(10),
		VisibilityTimeout:   types.Int64Ptr(30),
	})
	if err != nil {
		return nil, err
	}

	msgs := make([]*Message, len(results.Messages))
	for i, rawMsg := range results.Messages {
		var attrs map[string]string
		err := json.Unmarshal([]byte(types.StringValue(rawMsg.Body)), &attrs)
		if err != nil {
			return nil, err
		}

		if rawMeta, ok := attrs["NotificationMetadata"]; ok {
			var meta map[string]string
			err := json.Unmarshal([]byte(rawMeta), &meta)
			if err != nil {
				return nil, err
			}

			delete(attrs, "NotificationMetadata")
			for k, v := range meta {
				attrs[k] = v
			}
		}

		msgs[i] = &Message{
			ID:         types.StringValue(rawMsg.ReceiptHandle),
			Attributes: attrs,
		}
	}

	return msgs, nil
}

func DeleteFromQueue(client *aws.Client, queue, msgID string) error {
	svc := sqs.New(client.Session())
	url, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: types.StringPtr(queue),
	})
	if err != nil {
		return err
	}

	_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      url.QueueUrl,
		ReceiptHandle: types.StringPtr(msgID),
	})

	return err
}
