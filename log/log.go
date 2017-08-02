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
	ControlComponent    = "CONTROL"
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
	writer     io.Writer
	format     string
	verbosity  int
	component  string
	rotateFunc rotateFunc
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

func (lg *Logger) SetOutputWriter(writer io.Writer) {
	lg.writer = writer
	lg.rotateFunc = func() (string, error) {
		return "", fmt.Errorf("Cannot rotate arbitrary io.Writer")
	}
}

func (lg *Logger) SetOutputFile(filename string, logAppend bool, rotationStrategy string) error {
	w, rotate, err := newRotatingFile(filename, logAppend, rotationStrategy)
	if err == nil {
		lg.writer = w
		lg.rotateFunc = rotate
	}
	return err
}

func (lg *Logger) Rotate() (string, error) {
	return lg.rotateFunc()
}

func (lg *Logger) writelog(minVerbosity int, msg string) {
	if minVerbosity < 0 {
		panic("cannot set a minimum log verbosity that is less than 0")
	}

	if minVerbosity <= lg.verbosity {
		message := fmt.Sprintf("%v %v\n", time.Now().Format(lg.format), msg)
		lg.writer.Write([]byte(message))
	}
}

func NewLogger(verbosity VerbosityLevel) *Logger {
	noRotateFunc := func() (string, error) {
		return "", fmt.Errorf("cannot rotate logs without log path: use --logPath or in a config " +
			"file at 'systemLog.path'")
	}

	lg := &Logger{
		component:  defaultComponent,
		format:     defaultTimeFormat,
		rotateFunc: noRotateFunc,
		writer:     &stdErrWriter{},
	}

	lg.SetVerbosity(verbosity)

	return lg
}

func NewComponentLogger(component string, logger Logger) *Logger {
	lg := &Logger{
		component:  component,
		format:     logger.format,
		rotateFunc: logger.rotateFunc,
		verbosity:  logger.verbosity,
		writer:     logger.writer,
	}

	return lg
}

// stdErrWriter synchronizes writes to os.Stderr
type stdErrWriter struct{ sync.Mutex }

func (s *stdErrWriter) Write(b []byte) (int, error) {
	s.Lock()
	// TODO: maybe buffer writes here
	n, err := os.Stderr.Write(b)
	s.Unlock()
	return n, err
}

//// Log Writer Interface
type logWriter struct {
	logger       *Logger
	minVerbosity int
}

func (lgw *logWriter) Write(message []byte) (int, error) {
	lgw.logger.Log(lgw.minVerbosity, string(message))
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

func SetOutputWriter(writer io.Writer) {
	globalLogger.SetOutputWriter(writer)
}

func SetOutputFile(filename string, logAppend bool, rotationStrategy string) error {
	return globalLogger.SetOutputFile(filename, logAppend, rotationStrategy)
}

func SetDateFormat(dateFormat string) {
	globalLogger.SetDateFormat(dateFormat)
}

func Rotate() (string, error) {
	return globalLogger.Rotate()
}
