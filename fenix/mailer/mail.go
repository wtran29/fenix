// package to send mail
package mailer

import (
	"bytes"
	"fmt"
	"text/template"
)

type Mail struct {
	Domain      string
	Templates   string
	Host        string
	Port        string
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Jobs        chan Message
	Results     chan Result
	API         string
	APIKey      string
	APIUrl      string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Template    string
	Attachments []string
	Data        interface{}
}

type Result struct {
	Success bool
	Error   error
}

func (m *Mail) ListenForMail() {
	for {
		msg := <-m.Jobs
		err := m.Send(msg)
		if err != nil {
			m.Results <- Result{false, err}
		} else {
			m.Results <- Result{true, nil}
		}
	}
}

func (m *Mail) Send(msg Message) error {
	// TODO: use API or SMTP
	return m.SendSMTPMessage(msg)
}

func (m *Mail) SendSMTPMessage(msg Message) error {

	return nil
}

func (m *Mail) buildHTMLMsg(msg Message) (string, error) {
	tmplToRend := fmt.Sprintf("%s/%s.html.tmpl", m.Templates, msg.Template)

	tmpl, err := template.New("email-html").ParseFiles(tmplToRend)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = tmpl.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	formattedMsg := tpl.String()

	return formattedMsg, nil
}

func (m *Mail) buildPlainTextMsg(msg Message) (string, error) {
	tmplToRend := fmt.Sprintf("%s/%s.plain.tmpl", m.Templates, msg.Template)

	tmpl, err := template.New("email-html").ParseFiles(tmplToRend)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err = tmpl.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	plainMsg := tpl.String()

	return plainMsg, nil
}
