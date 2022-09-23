package svc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var (
	keys = map[string][]string{
		"DB":       []string{"DB_WRITE_HOST", "DB_READ_HOST", "DB_PORT", "DB_USER", "DB_PASS", "DB_NAME"},
		"Redis":    []string{"REDIS_WRITE_HOST", "REDIS_READ_HOST", "REDIS_PORT"},
		"API":      []string{"JWT_SECRET", "TOKEN_SECRET"},
		"Telegram": []string{"TELEGRAM_API_KEY", "TELEGRAM_INFO_CHAT_ID", "TELEGRAM_ERROR_CHAT_ID"},
		"Bittrex":  []string{"BITTREX_KEY", "BITTREX_SECRET"},
		"Kucoin":   []string{"KUCOIN_API_KEY", "KUCOIN_API_SECRET", "KUCOIN_API_PASSPHRASE"},
		"SES":      []string{"EMAIL_SENDER", "EMAIL_DOMAIN", "AWS_REGION"},
		"PDU":      []string{"PDU_ROCK_HOST", "PDU_HUT_HOST"},
		"Asana":    []string{"ASANA_PROJECT_ID", "ASANA_API_KEY"},
	}
)

func ParseSecrets(secretVar string) (map[string]string, error) {
	if len(secretVar) != 0 {
		raw := os.Getenv(secretVar)
		if len(raw) == 0 {
			return nil, fmt.Errorf("empty environment variable")
		}

		return godotenv.Unmarshal(raw)
	}

	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}

	return godotenv.Read(fmt.Sprintf("%s/.env", filepath.Dir(ex)))
}
