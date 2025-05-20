package log

import (
	"fmt"
	"time"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
)

func formattedTime() string {
	return time.Now().Format("Mon Jan 2006 15:04")
}

func Info(message string) {
	fmt.Printf("%s[INFO] %s %s\n", blue, formattedTime(), message+reset)
}

func Success(message string) {
	fmt.Printf("%s[SUCCESS] %s %s\n", green, formattedTime(), message+reset)
}

func Warning(message string) {
	fmt.Printf("%s[WARNING] %s %s\n", yellow, formattedTime(), message+reset)
}

func Error(message string) {
	fmt.Printf("%s[ERROR] %s %s\n", red, formattedTime(), message+reset)
}
