package export

import (
	"bytes"
	"encoding/csv"
	"strconv"
	"time"
)

/* formatting */

func formatBool(value bool) string {
	if value {
		return "true"
	}

	return "false"
}

func formatFloat64(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func formatTimestamp(value int64) string {
	return time.Unix(value, 0).Format(time.RFC1123)
}

/* writing */

func writeAsCSV(cols []string, rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	err := writer.Write(cols)
	if err != nil {
		return nil, err
	}

	err = writer.WriteAll(rows)
	if err != nil {
		return nil, err
	}

	writer.Flush()
	err = writer.Error()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
