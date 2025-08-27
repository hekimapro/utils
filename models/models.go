package models

// EncryptReturnType defines the structure for the encryption function’s return value
// Used to hold the encrypted payload in string format
type EncryptReturnType struct {
	Payload string // The encrypted data as a string (base64 or hex encoded)
}

// SMSRecipient represents a single recipient’s details in an SMS response
// Contains status and metadata for an SMS sent to a phone number
type ATSMSRecipient struct {
	StatusCode int    `json:"statusCode"` // HTTP status code for the SMS delivery
	Number     string `json:"number"`     // Recipient’s phone number
	Status     string `json:"status"`     // Status message (e.g., "Sent", "Queued")
	Cost       string `json:"cost"`       // Cost of sending the SMS
	MessageID  string `json:"messageId"`  // Unique ID for the SMS message
}

// SMSMessageData encapsulates the message and recipient data in an SMS response
// Aggregates the message content and list of recipients
type ATSMSMessageData struct {
	Message    string           `json:"Message"`    // The SMS message content
	Recipients []ATSMSRecipient `json:"Recipients"` // List of recipients and their statuses
}

// SMSResponse defines the structure of the Africa’s Talking API SMS response
// Wraps the message data and recipient information
type ATSMSResponse struct {
	SMSMessageData ATSMSMessageData `json:"SMSMessageData"` // The SMS message and recipient details
}

// SMSPayload defines the structure for sending a bulk SMS request
// Contains the necessary fields for the Africa’s Talking API
type ATSMSPayload struct {
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
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type BeemSMSRecipient struct {
	RecipientID string `json:"recipient_id"`
	PhoneNumber string `json:"dest_addr"`
}

type BeemSMSRequestBody struct {
	SourceAddr   string             `json:"source_addr"`
	ScheduleTime string             `json:"schedule_time,omitempty"` // Optional
	Encoding     string             `json:"encoding"`                // Default is "0"
	Message      string             `json:"message"`
	Recipients   []BeemSMSRecipient `json:"recipients"`
}

type BeemSMSResponse struct {
	Successful bool   `json:"successful"`
	RequestID  int    `json:"request_id"`
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Valid      int    `json:"valid"`
	Invalid    int    `json:"invalid"`
	Duplicates int    `json:"duplicates"`
}

type BeemSMSPayload struct {
	Message      string
	SenderName   string
	ScheduleTime string
	APIKey       string
	SecretKey    string
	Recipients   []BeemSMSRecipient
}

type BeemSMSDeliveryStatusResponse struct {
	DestAddr  string `json:"dest_addr"`
	Status    string `json:"status"`
	RequestID string `json:"request_id"`
}

type BeemSMSDeliveryStatusPayload struct {
	PhoneNumber string
	RequestID   int
	APIKey      string
	SecretKey   string
}

type DatabaseOptions struct {
	Username     string
	Password     string
	Host         string
	Port         string
	SSLMode      string // e.g., "disable", "require", "verify-full"
	DatabaseName string
}

type ContextKey string

type EncryptionConfig struct {
	EncryptionKey        string
	EncryptionType       string
	InitializationVector string
}
