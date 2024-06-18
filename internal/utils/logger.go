package utils // dnywonnt.me/alerts2incidents/internal/utils

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

func init() {
	// Define a boolean command-line flag 'debug' with default value false and a description using pflag.
	debugMode := pflag.BoolP("debug", "d", false, "Enable debug mode")

	// Parse the command-line flags.
	pflag.Parse()

	// If the 'debug' flag is set, update the log level to DebugLevel.
	if *debugMode {
		log.SetLevel(log.DebugLevel)
	}

	// Discard the default logging output and use a custom hook instead.
	log.SetOutput(io.Discard)

	// Add a custom logging hook to manage how logs are formatted and written.
	log.AddHook(&formatterHook{
		Writer: os.Stdout, // Set the output destination of the log to standard output.
		LogLevels: []log.Level{ // Define which log levels the hook should handle.
			log.InfoLevel,
			log.DebugLevel,
			log.WarnLevel,
			log.ErrorLevel,
			log.FatalLevel,
		},
		Formatter: &log.TextFormatter{ // Define the format of the log output.
			TimestampFormat: "2006-01-02 15:04:05", // Set the timestamp format.
			FullTimestamp:   true,                  // Enable full timestamp in the log output.
			ForceColors:     true,                  // Force colored output, even when not writing to a tty.
		},
	})
}

// formatterHook struct defines the structure for a custom logrus hook.
type formatterHook struct {
	Writer    io.Writer     // Writer where logs will be written.
	LogLevels []log.Level   // Log levels to be handled by this hook.
	Formatter log.Formatter // Formatter to format the log entries.
}

// Fire is called by logrus when a log entry is ready to be outputted.
func (hook *formatterHook) Fire(entry *log.Entry) error {
	line, err := hook.Formatter.Format(entry) // Format the log entry.
	if err != nil {
		return err // Return any error encountered during formatting.
	}

	_, err = hook.Writer.Write(line) // Write the formatted log entry to the writer.
	return err                       // Return any error encountered during writing.
}

// Levels returns the log levels that this hook is interested in.
func (hook *formatterHook) Levels() []log.Level {
	return hook.LogLevels // Return the log levels that the hook handles.
}
