package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Client struct {
	session *session.Session
}

func (c *Client) Session() *session.Session {
	return c.session
}

func NewSession(region, profile string) (*Client, error) {
	var cfg *aws.Config
	if len(profile) > 0 {
		cfg = &aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewSharedCredentials("", profile),
		}
	} else {
		cfg = &aws.Config{
			Region: aws.String(region),
		}
	}

	sess, err := session.NewSession(cfg)

	return &Client{sess}, err
}
