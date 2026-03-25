package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Setup initializes the global logrus logger from environment variables.
// LOG_LEVEL: debug | info | warn | error (default: info)
// LOG_FILE:  optional path to a log file (e.g. "logs/app.log")
// LOG_TO_STDOUT: true | false (default: false)
func Setup(appName string) {
	// Log level
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// JSON formatter
	logrus.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "msg",
		},
	})

	// Add "app" as a default field to every log entry
	logrus.AddHook(&appNameHook{name: appName})

	// Output writers
	writers := make([]io.Writer, 0)

	logToStdout := os.Getenv("LOG_TO_STDOUT")
	if logToStdout == "" || logToStdout == "true" {
		writers = append(writers, os.Stdout)
	}

	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		if err := os.MkdirAll(filepath(logFile), 0755); err == nil {
			f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				writers = append(writers, f)
			}
		}
	}

	if len(writers) > 0 {
		logrus.SetOutput(io.MultiWriter(writers...))
	}
}

// WithComponent returns an Entry pre-tagged with a component name.
// Usage: logger.WithComponent("prayer.service").Error("failed to create amen")
func WithComponent(component string) *logrus.Entry {
	return logrus.WithField("component", component)
}

// WithFields is a convenience re-export.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return logrus.WithFields(fields)
}

// filepath extracts the directory portion of a file path.
func filepath(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}

// appNameHook injects the "app" field into every log entry.
type appNameHook struct {
	name string
}

func (h *appNameHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *appNameHook) Fire(entry *logrus.Entry) error {
	entry.Data["app"] = h.name
	return nil
}
