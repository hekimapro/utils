package communication

import (
	"crypto/tls" // tls provides support for TLS configuration in network connections.
	"fmt"        // fmt provides formatting and printing functions.

	"github.com/hekimapro/utils/log"    // log provides colored logging utilities.
	"github.com/hekimapro/utils/models" // models contains data structures for email payloads.
	"gopkg.in/gomail.v2"                // gomail provides utilities for sending emails via SMTP.
)

// SendEmail sends an email using the provided SMTP server details and email content.
// Configures an email with sender, recipients, subject, body, and attachments.
// Connects to the SMTP server and sends the email, supporting TLS configuration.
// Returns an error if the email sending fails, otherwise nil.
func SendEmail(smtpHost string, smtpPort int, username, password string, InsecureSkipVerify bool, details models.EmailDetails) error {
	// Log the start of the email preparation process.
	log.Info("📤 Starting email preparation process")

	// Initialize a new email message.
	mail := gomail.NewMessage()

	// Set the email sender address.
	log.Info(fmt.Sprintf("📧 Setting email sender: %s", details.From))
	mail.SetHeader("From", details.From)

	// Set the email recipient addresses.
	log.Info(fmt.Sprintf("👥 Adding recipients: %v", details.To))
	mail.SetHeader("To", details.To...)

	// Set the email subject.
	log.Info(fmt.Sprintf("📝 Setting email subject: %s", details.Subject))
	mail.SetHeader("Subject", details.Subject)

	// Add plain text body if provided.
	if details.Text != "" {
		log.Info("📰 Adding plain text content to email")
		mail.SetBody("text/plain", details.Text)
	}

	// Add HTML body as an alternative if provided.
	if details.HTML != "" {
		log.Info("🌐 Adding HTML content to email")
		mail.AddAlternative("text/html", details.HTML)
	}

	// Attach files to the email if any are specified.
	for _, file := range details.Attachments {
		log.Info(fmt.Sprintf("📎 Attaching file: %s", file))
		mail.Attach(file)
	}

	// Create an SMTP dialer with the provided host, port, and credentials.
	log.Info(fmt.Sprintf("🔐 Creating SMTP dialer for host %s:%d", smtpHost, smtpPort))
	dialer := gomail.NewDialer(smtpHost, smtpPort, username, password)

	// Configure TLS settings, optionally skipping certificate verification.
	log.Warning(fmt.Sprintf("⚠️ TLS InsecureSkipVerify = %v", InsecureSkipVerify))
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: InsecureSkipVerify}

	// Attempt to connect to the SMTP server and send the email.
	log.Info("🚀 Attempting to send email...")
	if err := dialer.DialAndSend(mail); err != nil {
		// Log and return an error if sending fails.
		log.Error(fmt.Sprintf("❌ Failed to send email: %v", err))
		return fmt.Errorf("failed to send email: %v", err)
	}

	// Log successful email delivery.
	log.Success("✅ Email sent successfully!")
	return nil
}
