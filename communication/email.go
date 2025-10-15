package communication

import (
	"context"    // context provides support for cancellation and timeouts.
	"crypto/tls" // tls provides support for TLS configuration in network connections.
	"fmt"        // fmt provides formatting and printing functions.

	// io provides I/O interfaces for attachment handling.
	"mime" // mime provides MIME type detection.
	// net provides network utilities for custom timeouts.
	"os"            // os provides file system operations for attachments.
	"path/filepath" // filepath provides utilities for file path manipulation.
	"strings"       // strings provides utilities for string manipulation.
	"time"          // time provides functionality for timeouts and durations.

	"github.com/hekimapro/utils/helpers" // helpers provides utility functions.
	"github.com/hekimapro/utils/log"     // log provides colored logging utilities.
	"github.com/hekimapro/utils/models"  // models contains data structures for email payloads.
	"gopkg.in/gomail.v2"                 // gomail provides utilities for sending emails via SMTP.
)

// EmailConfig holds configuration for email sending operations.
type EmailConfig struct {
	SMTPHost           string        // SMTPHost is the SMTP server hostname
	SMTPPort           int           // SMTPPort is the SMTP server port
	Username           string        // Username for SMTP authentication
	Password           string        // Password for SMTP authentication
	InsecureSkipVerify bool          // InsecureSkipVerify controls TLS certificate verification
	Timeout            time.Duration // Timeout for email sending operations
	MaxRetries         int           // MaxRetries specifies maximum retry attempts
	RetryDelay         time.Duration // RetryDelay specifies delay between retries
}

// LoadEmailConfig loads email configuration with defaults.
func LoadEmailConfig() EmailConfig {
	return EmailConfig{
		SMTPHost:           helpers.GetENVValue("smtp host"),
		SMTPPort:           helpers.GetENVIntValue("smtp port", 587),
		Username:           helpers.GetENVValue("smtp username"),
		Password:           helpers.GetENVValue("smtp password"),
		InsecureSkipVerify: helpers.GetENVBoolValue("smtp insecure skip verify", false),
		Timeout:            time.Duration(helpers.GetENVIntValue("email timeout", 30)) * time.Second,
		MaxRetries:         helpers.GetENVIntValue("email max retries", 3),
		RetryDelay:         time.Duration(helpers.GetENVIntValue("email retry delay", 2)) * time.Second,
	}
}

// validateEmailDetails validates the email details before sending.
func validateEmailDetails(details models.EmailDetails) error {
	if details.From == "" {
		return helpers.CreateError("email 'From' address cannot be empty")
	}
	if len(details.To) == 0 {
		return helpers.CreateError("email must have at least one recipient")
	}
	if details.Subject == "" {
		return helpers.CreateError("email subject cannot be empty")
	}
	if details.Text == "" && details.HTML == "" {
		return helpers.CreateError("email must have either text or HTML content")
	}

	// Validate email addresses
	for _, recipient := range details.To {
		if !helpers.ValidateEmail(recipient) {
			return helpers.CreateErrorf("invalid email address: %s", recipient)
		}
	}

	return nil
}


// checkAttachmentExists verifies that attachment files exist and are readable.
func checkAttachmentExists(attachments []string) error {
	for _, file := range attachments {
		if file == "" {
			return helpers.CreateError("attachment path cannot be empty")
		}

		info, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				return helpers.CreateErrorf("attachment file does not exist: %s", file)
			}
			return helpers.WrapError(err, "failed to access attachment file")
		}

		if info.IsDir() {
			return helpers.CreateErrorf("attachment path is a directory, not a file: %s", file)
		}
	}
	return nil
}

// getMIMEType attempts to detect the MIME type of a file.
func getMIMEType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	// Common MIME type mappings
	mimeTypes := map[string]string{
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".zip":  "application/zip",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}

	// Fallback to system MIME type detection
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}

	return "application/octet-stream"
}

// createDialerWithTimeout creates a custom dialer with timeout support
func createDialerWithTimeout(config EmailConfig) *gomail.Dialer {
	dialer := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.Username, config.Password)

	// Configure TLS settings
	dialer.TLSConfig = &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
		ServerName:         config.SMTPHost,
	}

	// Since gomail.Dialer doesn't have a Timeout field, we rely on context for timeout control
	// The actual network operations will be wrapped in context-aware goroutines
	return dialer
}

// sendEmailWithContext is the internal implementation with context support.
func sendEmailWithContext(ctx context.Context, config EmailConfig, details models.EmailDetails) error {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return helpers.WrapError(ctx.Err(), "email sending cancelled before start")
	default:
		// Continue with email sending
	}

	// Log the start of the email preparation process.
	log.Info("📤 Starting email preparation process")

	// Validate email details
	if err := validateEmailDetails(details); err != nil {
		log.Error("❌ Email validation failed: " + err.Error())
		return err
	}

	// Check context cancellation after validation
	select {
	case <-ctx.Done():
		return helpers.WrapError(ctx.Err(), "email sending cancelled after validation")
	default:
		// Continue with email preparation
	}

	// Check attachments if any
	if len(details.Attachments) > 0 {
		if err := checkAttachmentExists(details.Attachments); err != nil {
			log.Error("❌ Attachment validation failed: " + err.Error())
			return err
		}
	}

	// Initialize a new email message.
	mail := gomail.NewMessage()

	// Set the email sender address.
	log.Info(fmt.Sprintf("📧 Setting email sender: %s", details.From))
	mail.SetHeader("From", details.From)

	// Set the email recipient addresses.
	log.Info(fmt.Sprintf("👥 Adding recipients: %v", details.To))
	mail.SetHeader("To", details.To...)

	// Set CC recipients if provided
	if len(details.CC) > 0 {
		log.Info(fmt.Sprintf("👥 Adding CC recipients: %v", details.CC))
		mail.SetHeader("Cc", details.CC...)
	}

	// Set BCC recipients if provided
	if len(details.BCC) > 0 {
		log.Info(fmt.Sprintf("👥 Adding BCC recipients: %v", details.BCC))
		mail.SetHeader("Bcc", details.BCC...)
	}

	// Set the email subject.
	log.Info(fmt.Sprintf("📝 Setting email subject: %s", details.Subject))
	mail.SetHeader("Subject", details.Subject)

	// Set reply-to if provided
	if details.ReplyTo != "" {
		log.Info(fmt.Sprintf("↩️ Setting reply-to: %s", details.ReplyTo))
		mail.SetHeader("Reply-To", details.ReplyTo)
	}

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

		// Get filename for attachment
		filename := filepath.Base(file)
		mimeType := getMIMEType(filename)

		mail.Attach(file, gomail.Rename(filename), gomail.SetHeader(map[string][]string{
			"Content-Type": {mimeType},
		}))
	}

	// Check context cancellation after message preparation
	select {
	case <-ctx.Done():
		return helpers.WrapError(ctx.Err(), "email sending cancelled after message preparation")
	default:
		// Continue with SMTP connection
	}

	// Create an SMTP dialer with the provided host, port, and credentials.
	log.Info(fmt.Sprintf("🔐 Creating SMTP dialer for host %s:%d", config.SMTPHost, config.SMTPPort))
	dialer := createDialerWithTimeout(config)

	// Configure TLS settings, optionally skipping certificate verification.
	log.Warning(fmt.Sprintf("⚠️ TLS InsecureSkipVerify = %v", config.InsecureSkipVerify))

	// Use a channel to handle the email sending with context
	sendDone := make(chan error, 1)

	go func() {
		// Attempt to connect to the SMTP server and send the email.
		log.Info("🚀 Attempting to send email...")
		sendDone <- dialer.DialAndSend(mail)
	}()

	// Wait for either the send to complete or context cancellation
	select {
	case <-ctx.Done():
		log.Warning("⚠️ Email sending cancelled or timed out")
		return helpers.WrapError(ctx.Err(), "email sending cancelled")
	case err := <-sendDone:
		if err != nil {
			// Log and return an error if sending fails.
			log.Error(fmt.Sprintf("❌ Failed to send email: %v", err))
			return helpers.WrapError(err, "failed to send email")
		}

		// Log successful email delivery.
		log.Success("✅ Email sent successfully!")
		return nil
	}
}

// SendEmail sends an email using the provided SMTP server details and email content.
// Configures an email with sender, recipients, subject, body, and attachments.
// Connects to the SMTP server and sends the email, supporting TLS configuration.
// Returns an error if the email sending fails, otherwise nil.
func SendEmail(smtpHost string, smtpPort int, username, password string, InsecureSkipVerify bool, details models.EmailDetails) error {
	// Load configuration with provided parameters
	config := EmailConfig{
		SMTPHost:           smtpHost,
		SMTPPort:           smtpPort,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: InsecureSkipVerify,
		Timeout:            30 * time.Second,
		MaxRetries:         0, // No retries for backward compatibility
		RetryDelay:         2 * time.Second,
	}

	// Create context with timeout for email operation
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	return sendEmailWithContext(ctx, config, details)
}

// SendEmailWithConfig sends an email using the provided configuration.
func SendEmailWithConfig(config EmailConfig, details models.EmailDetails) error {
	// Create context with timeout for email operation
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	return sendEmailWithContext(ctx, config, details)
}

// SendEmailWithRetry sends an email with retry logic for transient failures.
func SendEmailWithRetry(config EmailConfig, details models.EmailDetails) error {
	// Create context with timeout for retry operation (longer timeout)
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout*time.Duration(config.MaxRetries+1))
	defer cancel()

	return sendEmailWithRetryAndContext(ctx, config, details)
}

// sendEmailWithRetryAndContext is the internal implementation with context support and retry logic.
func sendEmailWithRetryAndContext(ctx context.Context, config EmailConfig, details models.EmailDetails) error {
	var lastError error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Check context cancellation before each attempt
		select {
		case <-ctx.Done():
			return helpers.WrapError(ctx.Err(), "email sending with retry cancelled")
		default:
			// Continue with attempt
		}

		// Log retry attempt if not the first attempt
		if attempt > 0 {
			log.Warning(fmt.Sprintf("🔄 Email retry attempt %d/%d", attempt, config.MaxRetries))
			// Wait before retry
			select {
			case <-ctx.Done():
				return helpers.WrapError(ctx.Err(), "email retry cancelled during delay")
			case <-time.After(config.RetryDelay * time.Duration(attempt)):
				// Continue with next attempt
			}
		}

		// Create a new context for this attempt with the original timeout
		attemptCtx, attemptCancel := context.WithTimeout(ctx, config.Timeout)

		// Attempt to send email
		err := sendEmailWithContext(attemptCtx, config, details)
		attemptCancel()

		if err == nil {
			// Success
			if attempt > 0 {
				log.Success(fmt.Sprintf("✅ Email sent successfully on attempt %d", attempt+1))
			}
			return nil
		}

		lastError = err

		// Check if error is retryable
		if !isRetryableEmailError(err) {
			log.Warning("⚠️ Non-retryable email error, not retrying")
			return err
		}

		// Log that we will retry
		if attempt < config.MaxRetries {
			log.Warning(fmt.Sprintf("⚠️ Retryable email error, will retry: %v", err))
		}
	}

	log.Error(fmt.Sprintf("❌ Email sending failed after %d attempts: %v", config.MaxRetries+1, lastError))
	return helpers.WrapError(lastError, "email sending failed after maximum retries")
}

// isRetryableEmailError checks if an email error is retryable.
func isRetryableEmailError(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := strings.ToLower(err.Error())

	// Retry on network-related errors
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"network error",
		"temporary failure",
		"too many connections",
		"rate limit",
		"quota exceeded",
		"busy",
		"try again",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errorMsg, pattern) {
			return true
		}
	}

	return false
}

// ValidateEmailAddress validates an email address format.
func ValidateEmailAddress(email string) bool {
	return helpers.ValidateEmail(email)
}

// SendTextEmail is a convenience function for sending plain text emails.
func SendTextEmail(smtpHost string, smtpPort int, username, password string, from string, to []string, subject, text string) error {
	details := models.EmailDetails{
		From:    from,
		To:      to,
		Subject: subject,
		Text:    text,
	}

	return SendEmail(smtpHost, smtpPort, username, password, false, details)
}

// SendHTMLEmail is a convenience function for sending HTML emails.
func SendHTMLEmail(smtpHost string, smtpPort int, username, password string, from string, to []string, subject, html string) error {
	details := models.EmailDetails{
		From:    from,
		To:      to,
		Subject: subject,
		HTML:    html,
	}

	return SendEmail(smtpHost, smtpPort, username, password, false, details)
}

// CreateEmailDetails creates a new EmailDetails struct with common fields.
func CreateEmailDetails(from string, to []string, subject, text, html string) models.EmailDetails {
	return models.EmailDetails{
		From:    from,
		To:      to,
		Subject: subject,
		Text:    text,
		HTML:    html,
	}
}

// AddAttachment adds a file attachment to email details.
func AddAttachment(details *models.EmailDetails, filePath string) error {
	if details == nil {
		return helpers.CreateError("email details cannot be nil")
	}

	if err := checkAttachmentExists([]string{filePath}); err != nil {
		return err
	}

	details.Attachments = append(details.Attachments, filePath)
	return nil
}

// AddRecipients adds multiple recipients to email details.
func AddRecipients(details *models.EmailDetails, recipients ...string) error {
	if details == nil {
		return helpers.CreateError("email details cannot be nil")
	}

	for _, recipient := range recipients {
		if !helpers.ValidateEmail(recipient) {
			return helpers.CreateErrorf("invalid email address: %s", recipient)
		}
		details.To = append(details.To, recipient)
	}

	return nil
}

// AddCC adds CC recipients to email details.
func AddCC(details *models.EmailDetails, recipients ...string) error {
	if details == nil {
		return helpers.CreateError("email details cannot be nil")
	}

	for _, recipient := range recipients {
		if !helpers.ValidateEmail(recipient) {
			return helpers.CreateErrorf("invalid email address: %s", recipient)
		}
		details.CC = append(details.CC, recipient)
	}

	return nil
}

// AddBCC adds BCC recipients to email details.
func AddBCC(details *models.EmailDetails, recipients ...string) error {
	if details == nil {
		return helpers.CreateError("email details cannot be nil")
	}

	for _, recipient := range recipients {
		if !helpers.ValidateEmail(recipient) {
			return helpers.CreateErrorf("invalid email address: %s", recipient)
		}
		details.BCC = append(details.BCC, recipient)
	}

	return nil
}
