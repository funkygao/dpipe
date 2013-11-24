package parser

import (
	"bytes"
	"fmt"
	"github.com/funkygao/alser/config"
	mail "github.com/funkygao/alser/sendmail"
	"github.com/funkygao/gotime"
	"time"
)

// TODO
type Alarm interface {
	String() string
}

func sendAlarmMailsLoop(conf *config.Config, mailBody *bytes.Buffer, bodyLines *int) {
	const mailTitlePrefix = "ALS Alarm"
	mailTo := conf.String("mail.guarded", "")
	if mailTo == "" {
		panic("empty mail.guarded")
	}

	mailSleep := conf.Int("mail.sleep_start", 120)
	backoffThreshold := conf.Int("mail.backoff_threshold", 10)
	bodyLineThreshold := conf.Int("line_threshold", 10)
	maxSleep, minSleep, sleepStep := conf.Int("mail.sleep_max", mailSleep*2),
		conf.Int("mail.sleep_min", mailSleep/2), conf.Int("mail.sleep_step", 5)
	for {
		select {
		case <-time.After(time.Second * time.Duration(mailSleep)):
			if *bodyLines >= bodyLineThreshold {
				go mail.Sendmail(mailTo, fmt.Sprintf("%s - %d", mailTitlePrefix, *bodyLines), mailBody.String())
				logger.Printf("alarm sent=> %s, sleep=%d\n", mailTo, mailSleep)

				// backoff sleep
				if *bodyLines >= backoffThreshold {
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

				mailBody.Reset()
				*bodyLines = 0
			}
		}
	}
}

func runSendAlarmsWatchdog(conf *config.Config) {
	var (
		bodyLines int
		mailBody  bytes.Buffer
	)

	go sendAlarmMailsLoop(conf, &mailBody, &bodyLines)

	for line := range chParserAlarm {
		if debug {
			logger.Printf("got alarm: %s\n", line)
		}

		mailBody.WriteString(fmt.Sprintf("%s %s\n",
			gotime.TsToString(int(time.Now().UTC().Unix())), line))
		bodyLines += 1
	}

}
