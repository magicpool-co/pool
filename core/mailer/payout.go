package mailer

import (
	_ "embed"
)

//go:embed templates/payout.html
var payoutTemplateData string
