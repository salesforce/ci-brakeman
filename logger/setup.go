package logger

var environ string

// Setup takes a DSN for sentry.io and creates the logging client
func Setup(env string) error {

	environ = env

	return nil
}
