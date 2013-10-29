package parser

import (
	"strings"
	"time"
)

// TODO
type Alarm interface {
	String() string
}

func runSendAlarmsWatchdog() {
	mailBody := ""

	for {
		select {
		case line, ok := <-chParserAlarm:
			if !ok {
				// chParserAlarm closed, this should never happen
				break
			}

			mailBody += line + "\n"

		case <-time.After(time.Second * 60):
			if mailBody != "" {
				mailBody = strings.TrimRight(mailBody, "\n")
				sendmailTo(emailRecipients, emailSubject, mailBody)
				logger.Printf("alarm mail sent: %s\n", emailRecipients)

				mailBody = ""
			}

		}
	}
}
