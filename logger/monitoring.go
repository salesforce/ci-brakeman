package logger

import (
	"fmt"
)

// Methods in this file are used to log monitoring data
// They provide easier ways to do structured logging

const (
	appName = "ci-brakeman"
)

// Debug write to the console/logstream
func Debug(s string) {
	fmt.Printf("DEBUG: %s\n", s)
}

// Debugf write to the console/logstream
func Debugf(s string, args ...interface{}) {
	fmt.Printf("DEBUG: %s", fmt.Sprintf(s, args))
}

// Info sends an info log
func Info(info string) {
	if environ == "staging" {
		fmt.Println("[EVENT] ", info)
	}
	//sentry.CaptureEvent()
}

// Infof sends an info log with parameters
func Infof(s string, v ...interface{}) {
	msg := fmt.Sprintf(s, v)
	fmt.Printf("[EVENT] %s", msg)
}

// Message sends a simple string  event
func Message(msg string) {
	fmt.Println(msg)
}

// Error sends an error log
func Error(err error) {
	fmt.Printf("[ERR] %v", err)
}

// Fatalf sends a fatal log with parameters
func Fatalf(s string, v ...interface{}) {
	//logFields().Fatalf(s, v...)
}

// Fatal sends a fatal log
func Fatal(args ...interface{}) {
	//logFields().Fatal(args...)
}

// Event sends a log event with custom fields
func Event() {

}

// CreateBreadcrumb creates a trace event
func CreateBreadcrumb(cat, msg string) {

	fmt.Printf("[EVENT][%s]: %s\n", cat, msg)
}
