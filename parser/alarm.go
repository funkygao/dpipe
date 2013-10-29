package parser

import (
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
				sendmailTo(emailRecipients, emailSubject, mailBody)
				logger.Println("alarm mail sent")
				mailBody = ""
			}

		}
	}
}
