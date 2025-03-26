package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

var (
	fileLogger    *log.Logger
	consoleLogger *log.Logger
)

func init() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}

	fileLogger = log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)
	consoleLogger = log.New(io.MultiWriter(os.Stdout), "", 0)
}

type logLevel int

const (
	infoLevel logLevel = iota
	warnLevel
	errorLevel
	fatalLevel
)

func logMessage(level logLevel, msg string, fields ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr, levelColor := getLevelInfo(level)
	formattedMsg := fmt.Sprintf(msg, fields...)

	fileLogger.Printf("[%s] %s - %s\n", levelStr, timestamp, formattedMsg)

	consoleLogger.Printf(
		"%s%s%s %s|%s %-7s %s|%s %s\n",
		colorCyan, timestamp, colorReset,
		colorReset, levelColor, levelStr, colorReset,
		colorReset, formattedMsg,
	)
}

func getLevelInfo(level logLevel) (string, string) {
	switch level {
	case infoLevel:
		return "INFO", colorGreen
	case warnLevel:
		return "WARN", colorYellow
	case errorLevel, fatalLevel:
		return "ERROR", colorRed
	default:
		return "UNKNOWN", colorReset
	}
}

func Info(msg string, fields ...interface{}) {
	logMessage(infoLevel, msg, fields...)
}

func Warn(msg string, fields ...interface{}) {
	logMessage(warnLevel, msg, fields...)
}

func Error(msg string, fields ...interface{}) {
	logMessage(errorLevel, msg, fields...)
}

func Fatal(msg string, fields ...interface{}) {
	logMessage(fatalLevel, msg, fields...)
	os.Exit(1)
}
