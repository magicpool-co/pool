package mailer

import (
	"text/template"

	"github.com/magicpool-co/pool/pkg/aws"
)

type Client struct {
	aws            *aws.Client
	workerTemplate *template.Template
	payoutTemplate *template.Template
}

func New(awsClient *aws.Client) (*Client, error) {
	workerTemplate, err := template.New("worker").Parse(workerTemplateData)
	if err != nil {
		return nil, err
	}

	payoutTemplate, err := template.New("payout").Parse(payoutTemplateData)
	if err != nil {
		return nil, err
	}

	client := &Client{
		aws:            awsClient,
		workerTemplate: workerTemplate,
		payoutTemplate: payoutTemplate,
	}

	return client, nil
}
