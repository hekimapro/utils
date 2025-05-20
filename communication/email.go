package communication

import (
	"crypto/tls"
	"fmt"

	"github.com/hekimapro/utils/log"
	"github.com/hekimapro/utils/models"
	"gopkg.in/gomail.v2"
)

// SendEmail sends an email using the provided SMTP server details and email content
// Configures an email with sender, recipients, subject, body, and attachments
// Connects to the SMTP server and sends the email, supporting TLS configuration
// Returns an error if the email sending fails, otherwise nil
func SendEmail(smtpHost string, smtpPort int, username, password string, InsecureSkipVerify bool, details models.EmailDetails) error {
	log.Info("Initializing email message")

	// Initialize a new gomail message instance for constructing the email
	mail := gomail.NewMessage()

	// Set the sender and recipient(s) in the email headers
	mail.SetHeader("From", details.From)
	mail.SetHeader("To", details.To...)

	log.Info(fmt.Sprintf("Setting email subject: %s", details.Subject))
	mail.SetHeader("Subject", details.Subject)

	// Add a plain text body if provided in the email details
	if details.Text != "" {
		log.Info("Adding plain text body to email")
		mail.SetBody("text/plain", details.Text)
	}

	// Add an HTML body as an alternative if provided in the email details
	if details.HTML != "" {
		log.Info("Adding HTML alternative body to email")
		mail.AddAlternative("text/html", details.HTML)
	}

	// Attach any files specified in the email details
	for _, file := range details.Attachments {
		log.Info(fmt.Sprintf("Attaching file: %s", file))
		mail.Attach(file)
	}

	// Create an SMTP dialer with the provided host, port, username, and password
	log.Info("Creating SMTP dialer")
	dialer := gomail.NewDialer(smtpHost, smtpPort, username, password)

	// Configure the TLS settings to allow insecure connections if specified
	log.Warning(fmt.Sprintf("TLS InsecureSkipVerify is set to %v", InsecureSkipVerify))
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: InsecureSkipVerify}

	// Attempt to connect to the SMTP server and send the email
	log.Info("Sending email...")
	if err := dialer.DialAndSend(mail); err != nil {
		log.Error(fmt.Sprintf("Email sending failed: %v", err))
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Success("Email sent successfully")
	// Return nil to indicate successful email sending
	return nil
}
