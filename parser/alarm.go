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
	mailSleep := conf.Int("mail.sleep_start", 120)
	maxSleep, minSleep, sleepStep := conf.Int("mail.sleep_max", mailSleep*2),
		conf.Int("mail.sleep_min", mailSleep/2), conf.Int("mail.sleep_step", 5)

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

		case <-time.After(time.Second * time.Duration(mailSleep)):
			if mailBody != "" {
				go mail.Sendmail(mailTo, fmt.Sprintf("%s - %d", mailTitlePrefix, bodyLines), mailBody)
				logger.Printf("alarm sent=> %s, sleep=%d\n", mailTo, mailSleep)

				// backoff sleep
				if bodyLines > 5 {
					// busy alarm
					mailSleep -= sleepStep
					if mailSleep < minSleep {
						mailSleep = minSleep
					}
				} else {
					// idle alarm
					mailSleep += sleepStep
					if mailSleep > maxSleep {
						mailSleep = maxSleep
					}
				}

				mailBody = ""
				bodyLines = 0
			}

		}
	}
}
