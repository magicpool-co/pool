package ses

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/types"
)

const (
	sender  = `MagicPool Notifications <no-reply@magicpool.co>`
	charset = "UTF-8"
)

func SendEmail(client *aws.Client, address, subject, body string) error {
	svc := ses.New(client.Session())
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				types.StringPtr(address),
			},
			CcAddresses: []*string{},
			BccAddresses: []*string{
				types.StringPtr("tug@sencha.dev"),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: types.StringPtr(charset),
					Data:    types.StringPtr(body),
				},
			},
			Subject: &ses.Content{
				Charset: types.StringPtr(charset),
				Data:    types.StringPtr(subject),
			},
		},
		Source: types.StringPtr(sender),
	}

	if _, err := svc.SendEmail(input); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				return fmt.Errorf("message rejected: %v", aerr)
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				return fmt.Errorf("mail from unverified domain: %v", aerr)
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				return fmt.Errorf("config set does not exist: %v", aerr)
			default:
				return aerr
			}
		} else {
			return err
		}
	}

	return nil
}
