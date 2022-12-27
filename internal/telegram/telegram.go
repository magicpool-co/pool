package telegram

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
)

type Client struct {
	Enabled     bool
	Key         string
	InfoChatID  int64
	ErrorChatID int64
}

func New(args map[string]string) (*Client, error) {
	var argKeys = []string{"TELEGRAM_API_KEY", "TELEGRAM_INFO_CHAT_ID", "TELEGRAM_ERROR_CHAT_ID"}
	for _, k := range argKeys {
		if _, ok := args[k]; !ok {
			return nil, fmt.Errorf("%s is a required argument", k)
		}
	}

	infoID, err := strconv.ParseInt(args["TELEGRAM_INFO_CHAT_ID"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse TELEGRAM_INFO_CHAT_ID")
	}

	errorID, err := strconv.ParseInt(args["TELEGRAM_ERROR_CHAT_ID"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse TELEGRAM_ERROR_CHAT_ID")
	}

	client := &Client{
		Enabled:     args["TELEGRAM_ENABLED"] != "false",
		Key:         args["TELEGRAM_API_KEY"],
		InfoChatID:  infoID,
		ErrorChatID: errorID,
	}

	return client, nil
}

type messageResponse struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
	Result      struct {
		MessageID uint64 `json:"message_id"`
		From      struct {
			ID        int64  `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"chat"`
		Date uint32 `json:"date"`
		Text string `json:"text"`
	} `json:"result"`
}

type StatusMessage struct {
	Chain       string
	RoundShares uint64
	Candidates  int
	Blocks      int
	Uncles      int
	Orphans     int
	Hashrate    float64
	Luck        float64
	Rewards     float64
}

func execPOST(url string, target interface{}) error {
	if httpReq, err := http.NewRequest("POST", url, nil); err != nil {
		return err
	} else {
		httpReq.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		if res, err := client.Do(httpReq); err != nil {
			return err
		} else {
			defer res.Body.Close()
			return json.NewDecoder(res.Body).Decode(&target)
		}
	}
}

func (t *Client) sendMessage(msg string, chatID int64) error {
	const parseMode = "MarkdownV2"
	const baseURL = "https://api.telegram.org"
	var replacer = strings.NewReplacer(
		"_", "\\_", "~", "\\~", ">", "\\>",
		"+", "\\+", "-", "\\-", "=", "\\=",
		"|", "\\|", "{", "\\{", "}", "\\}",
		".", "\\.", "!", "\\!",
	)

	if t.Enabled {
		msg = replacer.Replace(msg)
		msg = url.QueryEscape(msg)

		obj := new(messageResponse)
		url := fmt.Sprintf(`%s/bot%s/sendMessage?chat_id=%d&parse_mode=%s&text=%s`,
			baseURL, t.Key, chatID, parseMode, msg)

		if err := execPOST(url, obj); err != nil {
			return err
		} else if !obj.Ok {
			return fmt.Errorf("telegram responded with not ok: (%d) %s", obj.ErrorCode, obj.Description)
		}
	}

	return nil
}

/* error channel */

func (t *Client) SendFatal(err, app, env string) error {
	msg := fmt.Sprintf("Fatal error on `%s`:`%s` - `%s`", app, env, err)

	return t.sendMessage(msg, t.ErrorChatID)
}

func (t *Client) SendPanic(err, app, env string) error {
	msg := fmt.Sprintf("Panic on `%s`:`%s` - `%s`", app, env, err)

	return t.sendMessage(msg, t.ErrorChatID)
}

func (t *Client) NotifyNewHost(host, env string) error {
	msg := fmt.Sprintf("new stratum host on `%s` [%s]", host, env)

	return t.sendMessage(msg, t.ErrorChatID)
}

func (t *Client) NotifyNodeInstanceLaunched(chain, region string) error {
	msg := fmt.Sprintf("node instance launched for `%s` in %s", chain, region)

	return t.sendMessage(msg, t.ErrorChatID)
}

func (t *Client) NotifyNodeInstanceTerminated(chain, region string) error {
	msg := fmt.Sprintf("node instance terminated for `%s` in %s", chain, region)

	return t.sendMessage(msg, t.ErrorChatID)
}

/* info channel */

func (t *Client) NotifyNewBlockCandidate(chain, explorerURL string, height uint64, luck float64) error {
	msg := fmt.Sprintf("found %s block candidate with %.1f%% luck at [%d](%s)",
		strings.ToUpper(chain), luck, height, explorerURL)

	return t.sendMessage(msg, t.InfoChatID)
}

func (t *Client) NotifyInitiateExchangeBatch(id uint64) error {
	msg := fmt.Sprintf("initiated exchange batch %d", id)

	return t.sendMessage(msg, t.InfoChatID)
}

func (t *Client) NotifyInitiateDeposit(id uint64, chain string, value float64) error {
	msg := fmt.Sprintf("initated exchange deposit %d for %.4f %s", id, value, chain)

	return t.sendMessage(msg, t.InfoChatID)
}

func (t *Client) NotifyFinalizeDeposit(id uint64) error {
	msg := fmt.Sprintf("finalized exchange deposit %d", id)

	return t.sendMessage(msg, t.InfoChatID)
}

func (t *Client) NotifyInitiateTrade(id uint64, pathID, stageID int, market, direction string, value float64) error {
	msg := fmt.Sprintf("initated exchange trade %d \\(path: %d, stage: %d\\) for %s %.4f %s",
		id, pathID, stageID, direction, value, market)

	return t.sendMessage(msg, t.InfoChatID)
}

func (t *Client) NotifyFinalizeTrade(id uint64) error {
	msg := fmt.Sprintf("finalized exchange trade %d", id)

	return t.sendMessage(msg, t.InfoChatID)
}

func (t *Client) NotifyInitiateWithdrawal(id uint64, chain string, value float64) error {
	msg := fmt.Sprintf("initated exchange withdrawal %d for %.4f %s", id, value, chain)

	return t.sendMessage(msg, t.InfoChatID)
}

func (t *Client) NotifyFinalizeWithdrawal(id uint64) error {
	msg := fmt.Sprintf("finalized exchange withdrawal %d", id)

	return t.sendMessage(msg, t.InfoChatID)
}

func (t *Client) NotifyFinalizeExchangeBatch(id uint64) error {
	msg := fmt.Sprintf("completed exchange batch %d", id)

	return t.sendMessage(msg, t.InfoChatID)
}

func (t *Client) NotifyInitiatePayout(id uint64, chain, address, explorerURL string, value float64) error {
	msg := fmt.Sprintf("initated payout %d for %.4f %s to [%s](%s)", id, value, chain, address, explorerURL)

	return t.sendMessage(msg, t.InfoChatID)
}

func (t *Client) NotifyConfirmPayout(id uint64) error {
	msg := fmt.Sprintf("confirmed payout %d", id)

	return t.sendMessage(msg, t.InfoChatID)
}

func (t *Client) NotifyTransactionSent(id uint64, chain, txid, explorerURL string, value float64) error {
	msg := fmt.Sprintf("sent transaction %d for %.4f %s at [%s](%s)", id, value, chain, txid, explorerURL)

	return t.sendMessage(msg, t.InfoChatID)
}
