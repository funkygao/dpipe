package parser

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
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
Subject: {{.Subject}}
———————————-
{{.Body}}
———————————
EOF
`

	data := letterVar{to, subject, strings.TrimRight(body, "\n")}
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
