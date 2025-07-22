package env

import (
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env file, ignore error if file doesn't exist (optional)
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ .env file not found or failed to load: %v", err.Error())
	}
}

// toSnakeCase converts any input string to UPPER_SNAKE_CASE
func toSnakeCase(input string) string {
	s := strings.ReplaceAll(input, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	re := regexp.MustCompile(`([a-z])([A-Z])`)
	s = re.ReplaceAllString(s, "${1}_${2}")

	return strings.ToUpper(s)
}

// GetValue loads the environment variable value for a given key (case insensitive,
// converts input key to UPPER_SNAKE_CASE), including those loaded from .env file.
func GetValue(key string) string {
	snakeKey := toSnakeCase(key)
	return os.Getenv(snakeKey)
}
