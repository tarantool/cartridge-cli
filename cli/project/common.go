package project

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// RemoveTmpPath removes specified path if debug flag isn't set
// If path deletion fails, it warns
func RemoveTmpPath(path string, debug bool) {
	if debug {
		log.Warnf("%s is not removed due to debug mode", path)
		return
	}
	if err := os.RemoveAll(path); err != nil {
		log.Warnf("Failed to remove: %s", err)
	}
}
