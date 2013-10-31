package parser

import (
	"bytes"
	"os"
	"os/exec"
	"text/template"
)

func sendmailTo(to string, subject string, body string) {
	if to == "" || subject == "" || body == "" {
		logger.Println("empty mail params")
		return
	}

	type letterVar struct {
		To, Subject, Body string
	}

	const mailLetter = `From: ALS Guard <noreply@funplusgame.com>
To: {{.To}}
Subject: {{.Subject}}
MIME-Version: 1.0
Content-Type: text/html; charset="utf-8"
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
		logger.Println(err)
	}

	c1 := exec.Command("echo", wr.String())
	c2 := exec.Command("sendmail", "-t")
	c2.Stdin, _ = c1.StdoutPipe()
	c2.Stdout = os.Stdout
	c2.Start()
	c1.Run()
	c2.Wait()
}
