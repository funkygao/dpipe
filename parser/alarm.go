package parser

import (
	"bytes"
	"fmt"
	"github.com/funkygao/alser/rule"
	mail "github.com/funkygao/alser/sendmail"
	"github.com/funkygao/gotime"
	"time"
)

// TODO
type Alarm interface {
	String() string
}

func sendAlarmMailsLoop(ruleEngine *rule.RuleEngine, mailBody *bytes.Buffer, bodyLines *int) {
	const mailTitlePrefix = "ALS Alarm"
	mailTo := ruleEngine.String("mail.guarded", "")
	if mailTo == "" {
		panic("empty mail.guarded")
	}

	mailSleep := ruleEngine.Int("mail.sleep_start", 120)
	backoffThreshold := ruleEngine.Int("mail.backoff_threshold", 10)
	bodyLineThreshold := ruleEngine.Int("line_threshold", 10)
	maxSleep, minSleep, sleepStep := ruleEngine.Int("mail.sleep_max", mailSleep*2),
		ruleEngine.Int("mail.sleep_min", mailSleep/2), ruleEngine.Int("mail.sleep_step", 5)
	for {
		select {
		case <-time.After(time.Second * time.Duration(mailSleep)):
			if *bodyLines >= bodyLineThreshold {
				go mail.Sendmail(mailTo, fmt.Sprintf("%s - %d events", mailTitlePrefix, *bodyLines), mailBody.String())
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

func runSendAlarmsWatchdog(ruleEngine *rule.RuleEngine) {
	var (
		bodyLines int
		mailBody  bytes.Buffer
	)

	go sendAlarmMailsLoop(ruleEngine, &mailBody, &bodyLines)

	for line := range chParserAlarm {
		if debug {
			logger.Printf("got alarm: %s\n", line)
		}

		mailBody.WriteString(fmt.Sprintf("%s %s\n",
			gotime.TsToString(int(time.Now().UTC().Unix())), line))
		bodyLines += 1
	}

}
