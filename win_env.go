//go:build windows
// +build windows

package logger

import "os"

const DefaultLogPath = "log"

func chown(_ string, _ os.FileInfo) error {
	return nil
}
