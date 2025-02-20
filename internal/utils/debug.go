package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func DebugMsg(reqLvl int, curLvl int, logDir *string, msg string, args ...interface{}) {
	if curLvl < reqLvl {
		return
	}

	now := time.Now()
	timestamp := now.Format("15:04:05")

	fullMsg := fmt.Sprintf("[%d][%s] %s", reqLvl, timestamp, fmt.Sprintf(msg, args...))

	fmt.Println(fullMsg)

	// Check for log to file.
	if logDir != nil {
		// Format log file name.
		logFileName := fmt.Sprintf("%02d-%02d-%d.log", now.Month(), now.Day(), now.Year())
		logFilePath := filepath.Join(*logDir, logFileName)

		f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {
			fmt.Println("[DBG] Failed to open log file due to error: " + err.Error())

			return
		}

		defer f.Close()

		_, err = f.WriteString(fullMsg + "\n")
		if err != nil {
			fmt.Println("[DBG] Failed to write to log file due to error: " + err.Error())
		}
	}
}
