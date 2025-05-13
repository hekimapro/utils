package communication

import (
	"crypto/tls"
	"fmt"

	"github.com/hekimapro/utils/models"
	"gopkg.in/gomail.v2"
)

// SendEmail sends an email using the provided SMTP server details and email content
// Configures an email with sender, recipients, subject, body, and attachments
// Connects to the SMTP server and sends the email, supporting TLS configuration
// Returns an error if the email sending fails, otherwise nil
func SendEmail(smtpHost string, smtpPort int, username, password string, InsecureSkipVerify bool, details models.EmailDetails) error {
	// Initialize a new gomail message instance for constructing the email
	mail := gomail.NewMessage()

	// Set the sender and recipient(s) in the email headers
	mail.SetHeader("From", details.From)
	mail.SetHeader("To", details.To...)

	// Set the email subject in the header
	mail.SetHeader("Subject", details.Subject)

	// Add a plain text body if provided in the email details
	if details.Text != "" {
		mail.SetBody("text/plain", details.Text)
	}

	// Add an HTML body as an alternative if provided in the email details
	if details.HTML != "" {
		mail.AddAlternative("text/html", details.HTML)
	}

	// Attach any files specified in the email details
	for _, file := range details.Attachments {
		mail.Attach(file)
	}

	// Create an SMTP dialer with the provided host, port, username, and password
	dialer := gomail.NewDialer(smtpHost, smtpPort, username, password)

	// Configure the TLS settings to allow insecure connections if specified
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: InsecureSkipVerify}

	// Attempt to connect to the SMTP server and send the email
	if err := dialer.DialAndSend(mail); err != nil {
		// Return a wrapped error if the email sending fails
		return fmt.Errorf("failed to send email: %v", err)
	}

	// Return nil to indicate successful email sending
	return nil
}