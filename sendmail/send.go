/*
Use sendmail command instead of SMTP to send email.

You have to install a MTA on localhost before using this pkg.
*/
package parser

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"text/template"
)

func Sendmail(to string, subject string, body string) error {
	if to == "" || subject == "" || body == "" {
		return errors.New("empty mail params")
	}

	type letterVar struct {
		To, Subject, Body string
	}

	const mailLetter = `From: ALS Guard <noreply@funplusgame.com>
To: {{.To}}
Subject: {{.Subject}}
MIME-Version: 1.0
Content-Type: text/text; charset="utf-8"
Importance: High
X-Priority: 1 (Highest)
X-MSMail-Priority: High
——————————————————————
{{.Body}}
——————————————————————
`

	data := letterVar{to, subject, body}
	t := template.Must(template.New("letter").Parse(mailLetter))
	wr := new(bytes.Buffer)
	if err := t.Execute(wr, data); err != nil {
		return err
	}

	c1 := exec.Command("echo", wr.String())
	c2 := exec.Command("sendmail", "-t")
	c2.Stdin, _ = c1.StdoutPipe()
	c2.Stdout = os.Stdout
	c2.Start()
	c1.Run()
	c2.Wait()
	return nil
}
