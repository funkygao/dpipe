package parser

import (
	"fmt"
	mail "github.com/funkygao/alser/sendmail"
	"strings"
	"time"
)

// TODO
type Alarm interface {
	String() string
}

func runSendAlarmsWatchdog() {
	mailBody := ""
	bodyLines := 0

	for {
		select {
		case line, ok := <-chParserAlarm:
			if !ok {
				// chParserAlarm closed, this should never happen
				break
			}

			mailBody += line + "\n"
			bodyLines += 1

		case <-time.After(time.Second * 120):
			if mailBody != "" {
				mailBody = strings.TrimRight(mailBody, "\n")
				go mail.Sendmail("peng.gao@funplusgame.com", fmt.Sprintf("%s %d", "ALS", bodyLines), mailBody)
				logger.Printf("alarm sent=> %s\n", "peng.gao@funplusgame.com")

				mailBody = ""
				bodyLines = 0
			}

		}
	}
}
