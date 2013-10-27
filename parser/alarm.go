package parser

import (
	conf "github.com/daviddengcn/go-ljson-conf"
)

var (
	alarmEnabled                        = false
	emailSender, emailHost, emailPasswd string
)

// TODO
type Alarm interface {
	String() string
}

func init() {
	conf, err := conf.Load("conf/email.cf")
	if err == nil {
		emailSender = conf.String("sender", "")
		emailHost = conf.String("smtp_host", "")
		emailPasswd = conf.String("passwd", "")
		if verbose {
			logger.Printf("sender: %s smtp: %s\n", emailSender, emailHost)
		}
	}
}

func sendAlarm(to, subject, body string) {
	sendMail(emailSender, emailPasswd, emailHost, to, subject, body, false)
}
