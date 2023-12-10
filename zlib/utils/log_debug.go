//go:build debug
// +build debug

package utils

import (
	"log"
)

func Debug(format string, args ...interface{}) {
	log.Printf(format, args...)
}
