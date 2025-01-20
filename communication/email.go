package communication

import (
	"crypto/tls"
	"fmt"

	"github.com/hekimapro/utils/models"
	"gopkg.in/gomail.v2"
)

// SendEmail sends an email using the provided SMTP server details and email content.
// smtpHost: The SMTP server host (e.g., "smtp.gmail.com").
// smtpPort: The SMTP server port (e.g., 587 for TLS).
// username: The SMTP username (e.g., email address of sender).
// password: The SMTP password (or app-specific password).
// InsecureSkipVerify: A flag to allow insecure SSL connections (use with caution).
func SendEmail(smtpHost string, smtpPort int, username, password string, InsecureSkipVerify bool, details models.EmailDetails) error {
	// Create a new gomail message instance.
	mail := gomail.NewMessage()

	// Set sender and recipient(s) for the email.
	mail.SetHeader("From", details.From)
	mail.SetHeader("To", details.To...)

	// Set the email subject.
	mail.SetHeader("Subject", details.Subject)

	// Add the plain text body if provided.
	if details.Text != "" {
		mail.SetBody("text/plain", details.Text)
	}
	// Add the HTML body as an alternative if provided.
	if details.HTML != "" {
		mail.AddAlternative("text/html", details.HTML)
	}

	// Attach files if any are provided.
	for _, file := range details.Attachments {
		mail.Attach(file)
	}

	// Create a new SMTP dialer for connecting to the SMTP server.
	dialer := gomail.NewDialer(smtpHost, smtpPort, username, password)

	// Configure the dialer to allow insecure SSL connections (useful for development).
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: InsecureSkipVerify}

	// Attempt to send the email.
	if err := dialer.DialAndSend(mail); err != nil {
		// Return the error if the email sending fails.
		return fmt.Errorf("failed to send email: %v", err)
	}

	// Return nil if the email was sent successfully.
	return nil
}
