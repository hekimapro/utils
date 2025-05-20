package models

// EncryptReturnType defines the structure for the encryption function’s return value
// Used to hold the encrypted payload in string format
type EncryptReturnType struct {
	Payload string // The encrypted data as a string (base64 or hex encoded)
}

// SMSRecipient represents a single recipient’s details in an SMS response
// Contains status and metadata for an SMS sent to a phone number
type SMSRecipient struct {
	StatusCode int    `json:"statusCode"` // HTTP status code for the SMS delivery
	Number     string `json:"number"`     // Recipient’s phone number
	Status     string `json:"status"`     // Status message (e.g., "Sent", "Queued")
	Cost       string `json:"cost"`       // Cost of sending the SMS
	MessageID  string `json:"messageId"`  // Unique ID for the SMS message
}

// SMSMessageData encapsulates the message and recipient data in an SMS response
// Aggregates the message content and list of recipients
type SMSMessageData struct {
	Message    string         `json:"Message"`    // The SMS message content
	Recipients []SMSRecipient `json:"Recipients"` // List of recipients and their statuses
}

// SMSResponse defines the structure of the Africa’s Talking API SMS response
// Wraps the message data and recipient information
type SMSResponse struct {
	SMSMessageData SMSMessageData `json:"SMSMessageData"` // The SMS message and recipient details
}

// SMSPayload defines the structure for sending a bulk SMS request
// Contains the necessary fields for the Africa’s Talking API
type SMSPayload struct {
	Username     string   `json:"username"`     // Africa’s Talking account username
	Message      string   `json:"message"`      // The SMS message content
	SenderID     string   `json:"senderId"`     // Sender ID for the SMS
	PhoneNumbers []string `json:"phoneNumbers"` // List of recipient phone numbers
	ATAPIKey     string   `json:"apiKey"`       // API key for authentication
}

// EmailDetails defines the structure for email sending parameters
// Holds the sender, recipients, subject, body, and attachments for an email
type EmailDetails struct {
	From        string   // Sender email address
	To          []string // Recipient email addresses
	Subject     string   // Email subject
	Text        string   // Plain text message body
	HTML        string   // HTML message body
	Attachments []string // File paths for email attachments
}

// ServerResponse defines the structure for standardized JSON API responses
// Includes a success flag and a flexible message payload
type ServerResponse struct {
	Success bool        `json:"success"` // Indicates if the operation was successful
	Message interface{} `json:"message"` // The response payload (e.g., string, object)
}
