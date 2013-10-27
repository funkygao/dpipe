package parser

import (
	"net/smtp"
	"strings"
)

/*
user : peng.gao@funplusgame.com
password: xxxxx login smtp server password
host: smtp.exmail.qq.com:465 smtp.gmail.com:587
to: a@bar.com;b@163.com;c@foo.com.cn;...
mailtype: html or text
*/
func sendMail(user, password, host, to, subject, body, mailtype string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var contentType string
	if mailtype == "html" {
		contentType = "Content-Type: text/html; charset=UTF-8"
	} else {
		contentType = "Content-Type: text/plain; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + user + "<" + user + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	sendTo := strings.Split(to, ";")
	return smtp.SendMail(host, auth, user, sendTo, msg)
}
