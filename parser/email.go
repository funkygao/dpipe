package parser

import (
	"bytes"
	"os/exec"
	"text/template"
)

func sendmailTo(to string, subject string, body string) {
	if to == "" || subject == "" || body == "" {
		return
	}

	type letterVar struct {
		To, Subject, Body string
	}

	const mailLetter = `<<EOF
From: ALS Guard <noreply@funplusgame.com>
To: {{.To}}
Subject: {{.Body}}
———————————-
{{.Body}}
———————————
EOF
`

	data := letterVar{to, subject, body}
	t := template.Must(template.New("letter").Parse(mailLetter))
	wr := new(bytes.Buffer)
	if err := t.Execute(wr, data); err != nil {
		logger.Println(err)
	}
	cmd := exec.Command("sendmail", "-t", wr.String())
	var err error
	if err = cmd.Run(); err != nil {
		logger.Println(err)
	}

	if err = cmd.Wait(); err != nil {
		logger.Println(err)
	}

	logger.Println(wr.String())
}
