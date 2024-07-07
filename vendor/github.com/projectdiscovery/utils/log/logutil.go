package logutil

import (
	"io"
	"log"
	"os"
)

// DisableDefaultLogger disables the default logger.
func DisableDefaultLogger() {
	log.SetFlags(0)
	log.SetOutput(io.Discard)
}

// EnableDefaultLogger enables the default logger.
func EnableDefaultLogger() {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stderr)
}
