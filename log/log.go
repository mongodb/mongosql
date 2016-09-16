// Package log provides a utility to log timestamped messages to an io.Writer.
package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	Always = iota
	Info
	DebugLow
	DebugHigh
)

const (
	defaultTimeFormat = "2006-01-02T15:04:05.000-0700"
)

const (
	defaultComponent    = "MONGOSQLD"
	OptimizerComponent  = "OPTIMIZER"
	EvaluatorComponent  = "EXECUTOR"
	NetworkComponent    = "NETWORK"
	AlgebrizerComponent = "ALGEBRIZER"
)

// Logging Levels
const (
	Warning = "W" // Warning
	Error   = "F" // Fatal
)

var (
	level = map[int]string{
		Always:    "I", // Informational
		Info:      "I",
		DebugLow:  "D", // Debug
		DebugHigh: "D",
	}
)

type Logger struct {
	mutex     *sync.Mutex
	writer    io.Writer
	format    string
	verbosity int
	component string
}

type VerbosityLevel interface {
	Level() int
	IsQuiet() bool
}

func (lg *Logger) Errf(minVerbosity int, format string, a ...interface{}) {
	format = fmt.Sprintf(format, a...)
	format = fmt.Sprintf("%v %v %v", Error, lg.component, format)
	lg.writelog(minVerbosity, format)
}

func (lg *Logger) GetComponent() string {
	return lg.component
}

func (lg *Logger) log(msg string) {
	fmt.Fprintf(lg.writer, "%v %v\n", time.Now().Format(lg.format), msg)
}

func (lg *Logger) Log(minVerbosity int, format string) {
	format = fmt.Sprintf("%v %-10v %v", level[minVerbosity], lg.component, format)
	lg.writelog(minVerbosity, format)
}

func (lg *Logger) Logf(minVerbosity int, format string, a ...interface{}) {
	format = fmt.Sprintf(format, a...)
	format = fmt.Sprintf("%v %-10v %v", level[minVerbosity], lg.component, format)
	lg.writelog(minVerbosity, format)
}

func (lg *Logger) Warnf(minVerbosity int, format string, a ...interface{}) {
	format = fmt.Sprintf(format, a...)
	format = fmt.Sprintf("%v %-10v %v", Warning, lg.component, format)
	lg.writelog(minVerbosity, format)
}

func (lg *Logger) SetDateFormat(dateFormat string) {
	lg.format = dateFormat
}

func (lg *Logger) SetVerbosity(level VerbosityLevel) {
	if level == nil {
		lg.verbosity = 0
		return
	}

	if level.IsQuiet() {
		lg.verbosity = -1
	} else {
		lg.verbosity = level.Level()
	}
}

func (lg *Logger) SetWriter(writer io.Writer) {
	lg.writer = writer
}

func (lg *Logger) writelog(minVerbosity int, msg string) {
	if minVerbosity < 0 {
		panic("cannot set a minimum log verbosity that is less than 0")
	}

	if minVerbosity <= lg.verbosity {
		lg.mutex.Lock()
		fmt.Fprintf(lg.writer, "%v %v\n", time.Now().Format(lg.format), msg)
		lg.mutex.Unlock()
	}
}

func NewLogger(verbosity VerbosityLevel) *Logger {
	lg := &Logger{
		mutex:     &sync.Mutex{},
		writer:    os.Stderr,
		format:    defaultTimeFormat,
		component: defaultComponent,
	}
	lg.SetVerbosity(verbosity)
	return lg
}

func NewComponentLogger(component string, logger Logger) *Logger {
	lg := &Logger{
		mutex:     logger.mutex,
		writer:    logger.writer,
		format:    logger.format,
		verbosity: logger.verbosity,
		component: component,
	}
	return lg
}

//// Log Writer Interface
type logWriter struct {
	logger            *Logger
	minVerbosityosity int
}

func (lgw *logWriter) Write(message []byte) (int, error) {
	lgw.logger.Log(lgw.minVerbosityosity, string(message))
	return len(message), nil
}

// Writer returns an io.Writer that writes to the logger with
// the given verbosity
func (lg *Logger) Writer(minVerbosity int) io.Writer {
	return &logWriter{lg, minVerbosity}
}

//// Global Logging

var globalLogger *Logger

func init() {
	if globalLogger == nil {
		globalLogger = NewLogger(nil)
	}
}

func GlobalLogger() Logger {
	return *globalLogger
}

// IsInVerbosity returns true if the current verbosity level setting is
// greater than or equal to the given level.
func IsInVerbosity(minVerbosity int) bool {
	return minVerbosity <= globalLogger.verbosity
}

func Logf(minVerbosity int, format string, a ...interface{}) {
	globalLogger.Logf(minVerbosity, format, a...)
}

func Errf(minVerbosity int, format string, a ...interface{}) {
	globalLogger.Errf(minVerbosity, format, a...)
}

func Warnf(minVerbosity int, format string, a ...interface{}) {
	globalLogger.Warnf(minVerbosity, format, a...)
}

func Log(minVerbosity int, msg string) {
	globalLogger.Log(minVerbosity, msg)
}

func SetVerbosity(verbosity VerbosityLevel) {
	globalLogger.SetVerbosity(verbosity)
}

func SetWriter(writer io.Writer) {
	globalLogger.SetWriter(writer)
}

func SetDateFormat(dateFormat string) {
	globalLogger.SetDateFormat(dateFormat)
}

func Writer(minVerbosity int) io.Writer {
	return globalLogger.Writer(minVerbosity)
}
