package mailer

import (
	"bytes"
	_ "embed"
	"time"

	"github.com/magicpool-co/pool/pkg/aws/ses"
)

//go:embed templates/payout.html
var payoutTemplateData string

type payoutPage struct {
	Miner       string
	Date        string
	Value       string
	TxID        string
	ExplorerURL string
}

func (c *Client) generateEmailForPayout(templateData payoutPage) ([]byte, error) {
	var buf bytes.Buffer
	err := c.payoutTemplate.Execute(&buf, templateData)

	return buf.Bytes(), err
}

func (c *Client) SendEmailForPayout(emailAddress, miner, txid, explorerURL, value string, timestamp time.Time) error {
	subject := "A payout has been sent"

	templateData := payoutPage{
		Miner:       miner,
		Date:        timestamp.Format(time.RFC1123),
		Value:       value,
		TxID:        txid,
		ExplorerURL: explorerURL,
	}

	body, err := c.generateEmailForPayout(templateData)
	if err != nil {
		return err
	}

	return ses.SendEmail(c.aws, emailAddress, subject, string(body))
}
