package parser

import (
	"time"
)

// TODO
type Alarm interface {
	String() string
}

func sendEmailAlarm(to, subject, body string) {
	sendMail(emailSender, emailPasswd, emailHost, to, subject, body, false)
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
				sendEmailAlarm("peng.gao@funplusgame.com", "game error", mailBody)
				logger.Println("alarm mail sent")
				mailBody = ""
			}

		}
	}
}
