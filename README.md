# GO Utils

A comprehensive utility library for Go developers providing essential tools for server management, file handling, encryption, database operations, and more.

## Installation

```bash
go get github.com/hekimapro/utils
```

## Prerequisites

### For WebP Image Conversion
The image conversion functionality requires GCC to be installed on your system:

#### Windows
Install `GCC` or `MinGW` from [TDM-GCC](http://tdm-gcc.tdragon.net/download)

#### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install -y libwebp-dev build-essential
```

#### Verify Installation
```bash
# Check WebP library
dpkg -l | grep webp

# Check GCC version
gcc --version
```

#### macOS
```bash
brew install webp
```

## Packages Overview

### 1. Server (`server`)
Robust HTTP/HTTPS server with graceful shutdown and middleware support.

#### Features
- Automatic HTTP/HTTPS mode detection
- Graceful shutdown with configurable timeouts
- Health endpoint at `/health`
- Connection limiting
- Secure TLS configuration
- Middleware chaining

#### Usage
```go
import "github.com/hekimapro/utils/server"

func main() {
    router := http.NewServeMux()
    router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello World"))
    })

    // Start server with automatic TLS detection
    err := server.StartServer(router)
    if err != nil {
        log.Fatal(err)
    }
}
```

#### Environment Variables
```env
PORT=8080
SSL_KEY_PATH=/path/to/key.pem
SSL_CERT_PATH=/path/to/cert.pem
```

### 2. Scheduler (`scheduler`)
Background task scheduler with panic recovery and graceful shutdown.

#### Features
- Interval-based task execution
- Immediate execution option
- Panic recovery with stack traces
- Graceful shutdown support
- Execution metrics and monitoring

#### Usage
```go
import "github.com/hekimapro/utils/scheduler"

func myTask() {
    fmt.Println("Task executed at", time.Now())
}

// Run every 5 minutes, execute immediately
scheduler.RunFunctionAtInterval(myTask, 5*time.Minute, true)
```

### 3. Request (`request`)
HTTP client with retry logic, context support, and comprehensive error handling.

#### Features
- Automatic retry with exponential backoff
- Context support for cancellation
- JSON request/response handling
- Connection pooling
- Comprehensive error handling

#### Usage
```go
import "github.com/hekimapro/utils/request"

// GET request
response, err := request.Get("https://api.example.com/data", nil)
if err == nil {
    fmt.Println("Response:", string(response))
}

// POST request with body
payload := map[string]interface{}{"name": "John", "age": 30}
response, err := request.Post("https://api.example.com/users", payload, nil)

// With custom headers
headers := &request.Headers{
    "Authorization": "Bearer token",
    "X-Custom-Header": "value",
}
response, err := request.Get("https://api.example.com/protected", headers)
```

### 4. Log (`log`)
Structured, colored logging with multiple log levels and context support.

#### Features
- Colored console output
- Multiple log levels (DEBUG, INFO, SUCCESS, WARNING, ERROR)
- Timestamp formatting
- Caller information
- Structured logging support

#### Usage
```go
import "github.com/hekimapro/utils/log"

log.Info("Server starting on port 8080")
log.Success("Database connected successfully")
log.Warning("High memory usage detected")
log.Error("Failed to process request")
log.Debug("Debug information")

// Formatted logging
log.Infof("User %s logged in from %s", username, ipAddress)

// Structured logging
logger := log.WithFields(map[string]interface{}{
    "user_id": 12345,
    "action": "login",
})
logger.Info("User authentication")
```

#### Configuration
```go
// Enable debug logging
log.SetMinLevel(log.LevelDebug)

// Disable colors for production
log.DisableColors()

// Enable caller information
log.EnableCallerInfo()
```

### 5. Helpers (`helpers`)
Utility functions for common operations.

#### Features
- Environment variable handling
- JSON response helpers
- UUID operations
- File type detection
- Phone number normalization
- OTP generation
- Context utilities
- Pagination helpers

#### Usage
```go
import "github.com/hekimapro/utils/helpers"

// Environment variables
port := helpers.GetENVValue("port")
portWithDefault := helpers.GetENVIntValue("port", 8080)

// JSON responses
helpers.RespondWithJSON(w, http.StatusOK, data)
helpers.RespondWithError(w, http.StatusBadRequest, "Invalid input")

// UUID operations
userID := helpers.ConvertToUUID("a1b2c3d4-1234-5678-9101-abcdef123456")
if helpers.IsValidUUID(userID.String()) {
    // Valid UUID
}

// File operations
fileType := helpers.GetFileType("image.jpg") // Returns "image"

// Phone number normalization
phone := helpers.NormalizePhoneNumber("0755123456", false)

// OTP generation
otp, err := helpers.GenerateOTP() // 6-digit OTP

// Pagination
page, pageSize := helpers.GetPaginationParams(request)
offset := helpers.CalculateOffset(page, pageSize)
```

### 6. File (`file`)
File upload, conversion, and management utilities.

#### Features
- File upload with WebP conversion
- Secure filename generation
- Batch file operations
- Automatic rollback on failure
- File type validation
- Attachment handling

#### Usage
```go
import "github.com/hekimapro/utils/file"

// Upload single file
filename, err := file.UploadFile(fileReader, "image.jpg", "/uploads", true)

// Upload multiple files
filenames, err := file.UploadMultipleFiles(fileReaders, filenames, "/uploads", true)

// Delete files
err := file.DeleteFile("old-image.jpg", "/uploads")
err := file.DeleteMultipleFiles(filenames, "/uploads")

// Multipart form upload
result, err := file.UploadMultipartFile(fileHeader, "/uploads", true)

// File utilities
if file.FileExists("image.jpg", "/uploads") {
    info, _ := file.GetFileInfo("image.jpg", "/uploads")
    size, _ := file.GetFileSize("image.jpg", "/uploads")
}

// Cleanup old files
deleted, _ := file.CleanupOldFiles("/uploads", 24*time.Hour)
```

### 7. Encryption (`encryption`)
Cryptographic utilities for data encryption and password hashing.

#### Features
- AES-256 encryption/decryption
- Bcrypt password hashing
- Secure key generation
- Multiple encoding formats (Base64, Hex)

#### Usage
```go
import "github.com/hekimapro/utils/encryption"

// Data encryption
encrypted, err := encryption.Encrypt(sensitiveData)
if err == nil {
    decrypted, err := encryption.Decrypt(*encrypted)
}

// Password hashing
hashedPassword, err := encryption.CreateHash("myPassword123")
isValid := encryption.CompareWithHash(hashedPassword, "myPassword123")

// Custom cost hashing
hashedPassword, err := encryption.CreateHashWithCost("password", 12)

// Key generation
key, err := encryption.GenerateEncryptionKey(32) // 32 bytes for AES-256
iv, err := encryption.GenerateIV()
```

#### Environment Variables
```env
ENCRYPTION_TYPE=base64  # or "hex"
ENCRYPTION_KEY=your-32-byte-encryption-key
INITIALIZATION_VECTOR=your-16-byte-iv
```

### 8. Database (`database`)
PostgreSQL database utilities with connection pooling and transaction management.

#### Features
- Connection pooling with configurable limits
- Transaction management with retry logic
- Comprehensive error handling
- Connection health checks
- Query utilities with context support

#### Usage
```go
import "github.com/hekimapro/utils/database"

// Connect to database
db, err := database.ConnectToDatabase()
if err != nil {
    log.Fatal(err)
}
defer database.CloseDatabase(db)

// Execute transaction
err := database.Transaction(db, func(tx *sql.Tx) error {
    _, err := tx.Exec("INSERT INTO users (name) VALUES ($1)", "John")
    return err
})

// Transaction with retry
err := database.TransactionWithRetry(db, func(tx *sql.Tx) error {
    // Your database operations
}, 3) // Retry up to 3 times

// Health checks
if database.IsDatabaseConnected(db) {
    database.PingDatabase(db)
}

// Get statistics
database.PrintDatabaseStats(db)
```

#### Environment Variables
```env
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=mydb
DATABASE_USERNAME=user
DATABASE_PASSWORD=pass
DATABASE_SSL_MODE=disable
DATABASE_MAX_IDLE_CONNS=50
DATABASE_MAX_OPEN_CONNS=500
```

### 9. Communication (`communication`)
Email sending utilities with comprehensive error handling and retry logic.

#### Features
- SMTP email sending
- HTML and plain text support
- File attachments with MIME type detection
- Retry logic for transient failures
- Comprehensive validation

#### Usage
```go
import "github.com/hekimapro/utils/communication"

// Basic email
err := communication.SendEmail(
    "smtp.example.com",
    587,
    "username",
    "password",
    false,
    models.EmailDetails{
        From:    "sender@example.com",
        To:      []string{"recipient@example.com"},
        Subject: "Test Email",
        Text:    "Plain text content",
        HTML:    "<h1>HTML content</h1>",
    },
)

// With configuration
config := communication.LoadEmailConfig()
err := communication.SendEmailWithConfig(config, details)

// With retry logic
err := communication.SendEmailWithRetry(config, details)

// Convenience functions
err := communication.SendTextEmail(
    "smtp.example.com", 587, "user", "pass",
    "from@example.com", []string{"to@example.com"},
    "Subject", "Text content",
)
```

#### Environment Variables
```env
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=username
SMTP_PASSWORD=password
SMTP_INSECURE_SKIP_VERIFY=false
EMAIL_TIMEOUT=30
EMAIL_MAX_RETRIES=3
EMAIL_RETRY_DELAY=2
```

## Image Conversion (WebP)

The file package includes automatic WebP conversion for images:

```go
// Convert image to WebP format
convertedFile, err := file.CheckAndConvertFile(imageReader, "image.jpg")

// With custom quality settings
convertedFile, err := file.CheckAndConvertFileWithOptions(
    imageReader, "image.jpg", false, 0.8, // lossy, 80% quality
)

// Check if format is supported
if file.IsImageFormatSupported("photo.png") {
    // Can convert to WebP
}
```

**Supported formats:** JPG, JPEG, PNG, GIF

## Error Handling

All packages use consistent error handling with context:

```go
import "github.com/hekimapro/utils/helpers"

// Create errors
err := helpers.CreateError("error message")
err := helpers.CreateErrorf("error with %s", "format")

// Wrap errors
err = helpers.WrapError(originalErr, "additional context")
err = helpers.WrapErrorf(originalErr, "context with %s", "format")
```

## Context Support

All operations include context support for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Context is automatically used internally in all operations
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:
- Create an issue on GitHub
- Check existing documentation
- Review package examples

---

**Note:** This library is designed for production use with comprehensive error handling, context support, and security best practices. Always ensure proper configuration of environment variables for security-sensitive operations.