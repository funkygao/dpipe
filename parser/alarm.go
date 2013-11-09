package parser

import (
	"fmt"
	"github.com/funkygao/alser/config"
	mail "github.com/funkygao/alser/sendmail"
	"time"
)

// TODO
type Alarm interface {
	String() string
}

func runSendAlarmsWatchdog(conf *config.Config) {
	const mailTitlePrefix = "ALS Alarm"
	mailBody := ""
	bodyLines := 0
	mailTo := conf.String("mail.guarded", "")
	if mailTo == "" {
		panic("empty mail.guarded")
	}
	mailSleep := time.Duration(conf.Int("mail.sleep", 120))

	for {
		select {
		case line, ok := <-chParserAlarm:
			if !ok {
				// chParserAlarm closed, this should never happen
				break
			}

			if debug {
				logger.Printf("got alarm: %s\n", line)
			}

			mailBody += line + "\n"
			bodyLines += 1

		case <-time.After(time.Second * mailSleep):
			if mailBody != "" {
				go mail.Sendmail(mailTo, fmt.Sprintf("%s - %d", mailTitlePrefix, bodyLines), mailBody)
				logger.Printf("alarm sent=> %s\n", mailTo)

				mailBody = ""
				bodyLines = 0
			}

		}
	}
}
