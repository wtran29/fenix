// package to send mail
package mailer

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Templates   string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Jobs        chan Message
	Results     chan Result
	API         string
	APIKey      string
	APIDomain   string
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
	formattedMsg, err := m.buildHTMLMsg(msg)
	if err != nil {
		return err
	}

	plainMsg, err := m.buildPlainTextMsg(msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getMailEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)

	email.SetBody(mail.TextHTML, formattedMsg)
	email.AddAlternative(mail.TextPlain, plainMsg)

	if len(msg.Attachments) > 0 {
		for _, item := range msg.Attachments {
			email.AddAttachment(item)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mail) getMailEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSL
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
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
	formattedMsg, err = m.inlineCSS(formattedMsg)
	if err != nil {
		return "", err
	}

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

func (m *Mail) inlineCSS(s string) (string, error) {
	opt := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &opt)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", nil
	}

	return html, nil
}
