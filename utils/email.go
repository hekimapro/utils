package utils

import (
	"crypto/tls"

	"github.com/hekimapro/utils/models"
	"gopkg.in/gomail.v2"
)

// SendEmail sends an email with the provided details.
// smtpHost: The SMTP server host (e.g., "smtp.gmail.com").
// smtpPort: The SMTP server port (e.g., 587 for TLS).
// username: The SMTP username (e.g., email address of sender).
// password: The SMTP password (or app-specific password).
func SendEmail(smtpHost string, smtpPort int, username, password string, InsecureSkipVerify bool, details models.EmailDetails) error {
	mail := gomail.NewMessage()

	// Set sender and recipient(s).
	mail.SetHeader("From", details.From)
	mail.SetHeader("To", details.To...)

	// Set the subject.
	mail.SetHeader("Subject", details.Subject)

	// Set the email body (plain text and/or HTML).
	if details.Text != "" {
		mail.SetBody("text/plain", details.Text)
	}
	if details.HTML != "" {
		mail.AddAlternative("text/html", details.HTML)
	}

	// Attach files if any.
	for _, file := range details.Attachments {
		mail.Attach(file)
	}

	// Create a new dialer.
	dialer := gomail.NewDialer(smtpHost, smtpPort, username, password)

	// Allow insecure connections for development purposes (not recommended for production).
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: InsecureSkipVerify}

	// Send the email.
	if err := dialer.DialAndSend(mail); err != nil {
		return err
	}

	return nil
}
