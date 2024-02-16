package mailer

import (
	"bytes"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/magicpool-co/pool/pkg/aws/ses"
)

//go:embed templates/worker.html
var workerTemplateData string

type worker struct {
	Name        string
	LastShare   string
	IsEndOfList bool
}

type workerPage struct {
	Miner   string
	Workers []worker
}

func (c *Client) generateEmailForWorkers(
	miner string,
	workerIdx map[string]time.Time,
) ([]byte, error) {
	if len(workerIdx) == 0 {
		return nil, fmt.Errorf("empty worker list for template")
	}

	var i int
	workers := make([]worker, len(workerIdx))
	for name, lastShare := range workerIdx {
		workers[i] = worker{
			Name:        name,
			LastShare:   lastShare.Format(time.RFC1123),
			IsEndOfList: i == len(workerIdx)-1,
		}
		i++
	}

	templateData := workerPage{
		Miner:   miner,
		Workers: workers,
	}

	var buf bytes.Buffer
	err := c.workerTemplate.Execute(&buf, templateData)

	return buf.Bytes(), err
}

func (c *Client) SendEmailForWorkers(
	emailAddress, miner string,
	workerIdx map[string]time.Time,
) error {
	subject := "1 worker has gone down"
	if len(workerIdx) > 1 {
		subject = strconv.Itoa(len(workerIdx)) + " workers have gone down"
	}

	body, err := c.generateEmailForWorkers(miner, workerIdx)
	if err != nil {
		return err
	}

	return ses.SendEmail(c.aws, emailAddress, subject, string(body))
}
