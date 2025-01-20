package models

// EncryptReturnType defines the return type of the encryption function.
type EncryptReturnType struct {
	Payload string
}

type SMSRecipient struct {
	StatusCode int    `json:"statusCode"`
	Number     string `json:"number"`
	Status     string `json:"status"`
	Cost       string `json:"cost"`
	MessageID  string `json:"messageId"`
}

type SMSMessageData struct {
	Message    string         `json:"Message"`
	Recipients []SMSRecipient `json:"Recipients"`
}

type SMSResponse struct {
	SMSMessageData SMSMessageData `json:"SMSMessageData"`
}

type SMSPayload struct {
	Username     string   `json:"username"`
	Message      string   `json:"message"`
	SenderID     string   `json:"senderId"`
	PhoneNumbers []string `json:"phoneNumbers"`
	ATAPIKey     string   `json:"apiKey"`
}

// EmailDetails holds the necessary information for sending an email.
type EmailDetails struct {
	From       string   // Sender email address
	To         []string // Recipient email addresses
	Subject    string   // Email subject
	Text       string   // Plain text message
	HTML       string   // HTML message
	Attachments []string // Paths to files to attach
}

// ServerResponse represents the structure of the JSON response.
// Success indicates whether the operation was successful.
// Message contains the response message or payload (could be string, object, etc.).
type ServerResponse struct {
	Success bool        `json:"success"` // Indicates whether the operation was successful.
	Message interface{} `json:"message"` // Contains the response message or payload.
}
