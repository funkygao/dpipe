package plugins

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/funkygao/dpipe/engine"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var (
	normalizers = map[string]*regexp.Regexp{
		"digit":       regexp.MustCompile(`\d+`),
		"batch_token": regexp.MustCompile(`pre: .*; current: .*`),
	}
)

func indexName(project *engine.ConfProject, indexPattern string,
	date time.Time) (index string) {
	const (
		YM  = "@ym"
		YMW = "@ymw"
		YMD = "@ymd"

		INDEX_PREFIX = "fun_"
	)

	if strings.Contains(indexPattern, YM) {
		prefix := project.IndexPrefix
		fields := strings.SplitN(indexPattern, YM, 2)
		if fields[0] != "" {
			// e,g. rs@ym
			prefix = fields[0]
		}

		switch indexPattern {
		case YM:
			index = fmt.Sprintf("%s%s_%d_%02d", INDEX_PREFIX, prefix,
				date.Year(), int(date.Month()))
		case YMW:
			year, week := date.ISOWeek()
			index = fmt.Sprintf("%s%s_%d_w%02d", INDEX_PREFIX, prefix,
				year, week)
		case YMD:
			index = fmt.Sprintf("%s%s_%d_%02d_%02d", INDEX_PREFIX, prefix,
				date.Year(), int(date.Month()), date.Day())
		}

		return
	}

	index = INDEX_PREFIX + indexPattern

	return
}

// Use sendmail command instead of SMTP to send email
func Sendmail(to string, subject string, body string) error {
	if to == "" || subject == "" || body == "" {
		return errors.New("empty mail params")
	}

	type letterVar struct {
		To, Subject, Body string
	}

	const mailLetter = `From: dpipe <noreply@funplusgame.com>
To: {{.To}}
Subject: {{.Subject}}
Importance: High
X-Priority: 1 (Highest)
X-MSMail-Priority: High
——————————————————————
{{.Body}}
====
dpiped
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
