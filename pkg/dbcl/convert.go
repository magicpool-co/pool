package dbcl

import (
	"fmt"
	"time"
)

func ConvertDurationToInterval(duration time.Duration) string {
	ns := int64(duration)

	interval := "1 SECOND"
	if amt := ns / int64(time.Hour*24*30); amt > 0 {
		interval = fmt.Sprintf("%d MONTH", amt)
	} else if amt := ns / int64(time.Hour*24); amt > 0 {
		interval = fmt.Sprintf("%d DAY", amt)
	} else if amt := ns / int64(time.Hour); amt > 0 {
		interval = fmt.Sprintf("%d HOUR", amt)
	} else if amt := ns / int64(time.Minute); amt > 0 {
		interval = fmt.Sprintf("%d MINUTE", amt)
	} else if amt := ns / int64(time.Second); amt > 0 {
		interval = fmt.Sprintf("%d SECOND", amt)
	}

	return "INTERVAL " + interval
}
