package tag

import (
	"os"
	"fmt"
	"log"
	"github.com/hashicorp/logutils"
)

// alias for logging to console
// TODO: Improve logging -> Preferably to a file.

func init() {
	filter := &logutils.LevelFilter{
		Levels: []logutils.LogLevel{"WARN", "FATAL", "INFO"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer: os.Stderr,
	}
	log.SetOutput(filter)
}

func Info (message string) {
	log.Print(fmt.Sprintf("[INFO]: %s", message))
}

func Fatal (message string) {
	log.Fatal(fmt.Sprintf("[FATAL]: %s", message))
}

func Warn (message string) {
	log.Print(fmt.Sprintf("[WARN]: %s", message))
}
