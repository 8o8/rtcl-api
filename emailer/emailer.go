/*
	Package Email sends emails using Sendgrid API. At this stage can only do single emails with one attachemnt.
*/
package emailer

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"log"
)

const SENGRID_API_KEY = "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"

type Email struct {
	APIKey       string
	FromName     string
	FromEmail    string
	ToEmail      string
	ToName       string
	Subject      string
	PlainContent string
	HTMLContent  string
	Attachments  []Attachment
}

type Attachment struct {
	MIMEType      string
	FileName      string
	Base64Content string
}

// New sets the APIKey and returns a pointer to an Email
func New() *Email {
	return &Email{
		APIKey: SENGRID_API_KEY,
	}
}

func (e Email) Send() error {

	message := prepare(e)

	for _, a := range e.Attachments {
		attach(a, message)
	}

	client := sendgrid.NewSendClient(e.APIKey)
	_, err := client.Send(message)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (a Attachment) isOK() bool {
	return a.Base64Content != "" && a.FileName != "" && a.MIMEType != ""
}

func prepare(e Email) *mail.SGMailV3 {
	from := mail.NewEmail(e.FromName, e.FromEmail)
	subject := e.Subject
	to := mail.NewEmail(e.ToName, e.ToEmail)
	plainTextContent := e.PlainContent
	htmlContent := e.HTMLContent
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	return message
}

func attach(a Attachment, message *mail.SGMailV3) {
	if a.isOK() {
		message.AddAttachment(newAttachment(a))
	}
}

func newAttachment(a Attachment) *mail.Attachment {
	ma := mail.NewAttachment()
	ma.SetContent(a.Base64Content)
	ma.SetType(a.MIMEType)
	ma.SetFilename(a.FileName)
	ma.SetDisposition("attachment") // no "inline" for now
	// ma.SetContentID("Attachment...") // used for inline attachments
	return ma
}
