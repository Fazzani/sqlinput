package models

import (
	"fmt"
	"os"
)

type logLevel string

const (
	LogDebug logLevel = "DEBUG"
	LogInfo           = "INFO"
	LogWarn           = "WARN"
	LogError          = "ERROR"
	LogFatal          = "FATAL"
)

// Logf log Format for splunk
func Logf(level logLevel, format string, msg ...interface{}) (n int, err error) {
	return fmt.Fprintf(os.Stderr, "%s %s\n", level, fmt.Sprintf(format, msg...))
}
