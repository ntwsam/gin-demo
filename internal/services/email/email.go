package email

import (
	"net/smtp"
	"os"
	"strings"
)

type EmailData struct {
	To      []string
	Subject string
	Body    string
}

func SendEmail(data EmailData) error {
	from := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	to := strings.Join(data.To, ",") // 🐳 รวม email หลายๆอัน

	message := []byte("Subject: " + data.Subject + "\r\n" +
		"From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"\r\n" + data.Body)

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, data.To, message)
	if err != nil {
		return err
	}
	return nil
}
