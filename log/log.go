// Package log provides a utility to log timestamped messages to an io.Writer.
package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/10gen/sqlproxy/internal/util"
)

const (
	defaultTimeFormat        = "2006-01-02T15:04:05.000-0700"
	flushInterval            = 400 * time.Millisecond
	bufferSizeFlushThreshold = 4096
	bufferSizeLimit          = 8192
)

// Constants for various mongosqld logging components
const (
	ControlComponent    = "CONTROL"
	OptimizerComponent  = "OPTIMIZER"
	EvaluatorComponent  = "EXECUTOR"
	NetworkComponent    = "NETWORK"
	AlgebrizerComponent = "ALGEBRIZER"
	LoggerComponent     = "LOGGER"
	SamplerComponent    = "SAMPLER"
	MongodrdlComponent  = "MONGODRDL"
)

// Verbosity is an enum representing logging verbosity levels.
type Verbosity int

const (
	// Quiet instructs the logger to never log any output.
	Quiet Verbosity = -1

	// Always is for messages that notify the user of basic mongosqld events and state changes.
	Always Verbosity = 0

	// Admin is for messages that would be useful/understandable for mongosqld admins.
	Admin Verbosity = 1

	// Dev is for messages targeted primarily at MongoDB developers, TSEs, etc.
	Dev Verbosity = 2
)

// Severity is an enum representing log message severities.
type Severity string

const (
	// Info is for messages that are understandable/useful for the target audience
	// of the message's verbosity level without any additional context (beyond an
	// understanding of mongosqld).
	Info Severity = "I"

	// Debug is for informational messages that require additional context to
	// be useful or understandable.
	Debug Severity = "D"

	// Warn is for unusual or unexpected mongosqld behavior/errors from which
	// mongosqld will automatically recover.
	Warn Severity = "W"

	// Error is for errors fatal to an individual operation, but not to
	// mongosqld itself.
	Error Severity = "E"

	// Fatal is for errors from which mongosqld cannot recover.
	// It should be used sparingly, and only in cases where something
	// is seriously broken.
	Fatal Severity = "F"
)

func loggingComponentLogger() *Logger {
	return NewComponentLogger(LoggerComponent, GlobalLogger())
}

// newLine gets the proper line ending for the current OS.
// OS will not change mid-run, so we set our logging line endings once, on package init.
// We need a function to do this in go because there is no conditional expression
// for setting a global variable
func newLine() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

var (
	// NewLine is the actual newline string to use when logging.
	NewLine = newLine()
)

// A Logger provides an API for writing log messages with various severities and
// verbosities.
type Logger struct {
	buffer     *writeBuffer
	verbosity  Verbosity
	component  string
	rotateFunc rotateFunc
}

// GetComponent returns this logger's component.
func (lg *Logger) GetComponent() string {
	return lg.component
}

// Infof writes the provided log message at the specified verbosity with severity Info.
func (lg *Logger) Infof(minVerbosity Verbosity, format string, a ...interface{}) {
	format = fmt.Sprintf(format, a...)
	lg.writelog(minVerbosity, Info, format)
}

// Debugf writes the provided log message at the specified verbosity with severity Debug.
func (lg *Logger) Debugf(minVerbosity Verbosity, format string, a ...interface{}) {
	format = fmt.Sprintf(format, a...)
	lg.writelog(minVerbosity, Debug, format)
}

// Warnf writes the provided log message at the specified verbosity with severity Warn.
func (lg *Logger) Warnf(minVerbosity Verbosity, format string, a ...interface{}) {
	format = fmt.Sprintf(format, a...)
	lg.writelog(minVerbosity, Warn, format)
}

// Errf writes the provided log message at the specified verbosity with severity Err.
func (lg *Logger) Errf(minVerbosity Verbosity, format string, a ...interface{}) {
	format = fmt.Sprintf(format, a...)
	lg.writelog(minVerbosity, Error, format)
}

// Fatalf writes the provided log message at the specified verbosity
// with severity Fatal and then panics.
func (lg *Logger) Fatalf(minVerbosity Verbosity, format string, a ...interface{}) {
	format = fmt.Sprintf(format, a...)
	lg.writelog(minVerbosity, Fatal, format)
}

// SetDateFormat sets the date format that this logger should use to write timestamps.
func (lg *Logger) SetDateFormat(dateFormat string) {
	lg.buffer.setDateFormat(dateFormat)
}

// SetVerbosity sets the maximum verbosity level for which
// messages should be logged.
func (lg *Logger) SetVerbosity(level Verbosity) {
	switch level {
	case Quiet, Always, Admin, Dev:
		lg.verbosity = level
	default:
		if level > Dev {
			Warnf(Always, "log verbosity level %d does not exist; using verbosity Dev", level)
			lg.verbosity = Dev
		} else {
			Warnf(Always, "log verbosity level %d does not exist; using verbosity Always", level)
			lg.verbosity = Always
		}
	}
}

// SetOutputWriter provides a writer to which this logger should write its messages.
func (lg *Logger) SetOutputWriter(writer io.Writer) {
	lg.buffer.setWriter(writer)
	lg.rotateFunc = noRotateFunc
}

// SetOutputFile instructs the logger to write its log messages to the specified file.
func (lg *Logger) SetOutputFile(filename string, logAppend bool, strategy RotationStrategy) error {
	w, err := newRotatingFile(filename, logAppend, strategy)
	if err == nil {
		lg.buffer.setWriter(w)
		lg.rotateFunc = w.rotate
	}
	return err
}

// Flush requests that the logger's underlying buffer flush its write buffer.
func (lg *Logger) Flush() {
	lg.buffer.requestFlush(true)
}

// Rotate causes the logger to rotate its output file, if possible.
// If rotation is successful, the location of the rotated log file is returned.
// If rotation fails, an error is returned.
func (lg *Logger) Rotate() (string, error) {
	return lg.rotateFunc()
}

func (lg *Logger) writelog(minVerbosity Verbosity, severity Severity, msg string) {
	if minVerbosity < 0 {
		panic("cannot set a minimum log verbosity that is less than 0")
	}

	msg = fmt.Sprintf("%v %-10v %v", severity, lg.component, msg)

	if minVerbosity <= lg.verbosity {
		lg.buffer.writeMessage(msg)
	}
}

func noRotateFunc() (string, error) {
	return "", fmt.Errorf("cannot rotate logs without log path: use --logPath or in a config " +
		"file at 'systemLog.path'")
}

// NewLogger returns a new logger with the specified verbosity.
func NewLogger(verbosity Verbosity) *Logger {
	lg := &Logger{
		buffer:     newWriteBuffer(os.Stderr),
		component:  ControlComponent,
		rotateFunc: noRotateFunc,
	}

	lg.SetVerbosity(verbosity)

	return lg
}

// NewComponentLogger returns a new logger that will write messages to the
// provided parent logger with the specified component.
func NewComponentLogger(component string, logger *Logger) *Logger {
	lg := &Logger{
		buffer:     logger.buffer,
		component:  component,
		rotateFunc: logger.rotateFunc,
		verbosity:  logger.verbosity,
	}

	return lg
}

type writeBuffer struct {
	tmp     []byte
	buf     []byte
	bufLock sync.Mutex

	threshold int
	limit     int

	writer     io.Writer
	writerLock sync.Mutex

	flushChan chan chan struct{}

	format string
}

func newWriteBuffer(writer io.Writer) *writeBuffer {
	flushChan := make(chan chan struct{}, 1)
	w := &writeBuffer{
		threshold: bufferSizeFlushThreshold,
		limit:     bufferSizeLimit,
		writer:    writer,
		flushChan: flushChan,
		format:    defaultTimeFormat,
	}

	util.PanicSafeGo(func() {
		for {
			select {
			case <-time.After(flushInterval):
				w.flush()
			case done := <-flushChan:
				w.flush()
				done <- struct{}{}
			}
		}
	}, func(err interface{}) {
		loggingComponentLogger().Fatalf(Always, "error flushing logs: %v", err)
		panic(err)
	})

	return w
}

func (w *writeBuffer) requestFlush(wait bool) {
	// create a channel for callback after completed flush
	done := make(chan struct{}, 1)

	if wait {
		// we need to make sure this exact flush request gets serviced
		w.flushChan <- done
		<-done
	} else {
		// we just care that a flush request gets submitted
		select {
		case w.flushChan <- done:
		default:
		}
	}
}

func (w *writeBuffer) writeMessage(str string) {
	w.bufLock.Lock()
	bufLen := len(w.buf) + len(str)
	w.bufLock.Unlock()

	if bufLen > w.limit {
		w.requestFlush(true)
	} else if bufLen > w.threshold {
		w.requestFlush(false)
	}

	w.bufLock.Lock()
	msgBytes := []byte(fmt.Sprintf("%v %v%s", time.Now().Format(w.format), str, NewLine))
	w.buf = append(w.buf, msgBytes...)
	w.bufLock.Unlock()
}

// flush is _not_ thread safe.
// concurrent use will result in data races on w.tmp
func (w *writeBuffer) flush() {

	w.bufLock.Lock()

	bufLen := len(w.buf)
	if bufLen == 0 {
		w.bufLock.Unlock()
		return
	}

	// resize tmp slice
	if cap(w.tmp) < bufLen {
		// allocate if we need more capacity
		w.tmp = make([]byte, bufLen)
	} else {
		// otherwise just reslice
		w.tmp = w.tmp[:bufLen]
	}

	// copy currently buffered data into tmp slice
	copy(w.tmp, w.buf)

	// truncate the buffer
	w.buf = w.buf[:0]

	w.bufLock.Unlock()

	// get the current writer
	w.writerLock.Lock()
	writer := w.writer
	w.writerLock.Unlock()

	// write the flushed data to the writer
	_, err := writer.Write(w.tmp)
	if err != nil {
		panic(err)
	}
}

func (w *writeBuffer) setWriter(writer io.Writer) {
	w.writerLock.Lock()
	w.writer = writer
	w.writerLock.Unlock()
}

func (w *writeBuffer) setDateFormat(dateFormat string) {
	w.bufLock.Lock()
	w.format = dateFormat
	w.bufLock.Unlock()
}

//// Global Logging

var globalLogger *Logger

func init() {
	if globalLogger == nil {
		globalLogger = NewLogger(Always)
	}
}

// GlobalLogger returns the global logger instance.
func GlobalLogger() *Logger {
	return globalLogger
}

// Infof writes a message with severity Info to the global logger.
func Infof(minVerbosity Verbosity, format string, a ...interface{}) {
	globalLogger.Infof(minVerbosity, format, a...)
}

// Debugf writes a message with severity Debug to the global logger.
func Debugf(minVerbosity Verbosity, format string, a ...interface{}) {
	globalLogger.Debugf(minVerbosity, format, a...)
}

// Warnf writes a message with severity Warn to the global logger.
func Warnf(minVerbosity Verbosity, format string, a ...interface{}) {
	globalLogger.Warnf(minVerbosity, format, a...)
}

// Errf writes a message with severity Err to the global logger.
func Errf(minVerbosity Verbosity, format string, a ...interface{}) {
	globalLogger.Errf(minVerbosity, format, a...)
}

// Fatalf writes a message with severity Fatal to the global logger and panics.
func Fatalf(minVerbosity Verbosity, format string, a ...interface{}) {
	globalLogger.Fatalf(minVerbosity, format, a...)
}

// SetVerbosity sets the verbosity of the global logger.
func SetVerbosity(verbosity Verbosity) {
	globalLogger.SetVerbosity(verbosity)
}

// SetOutputWriter sets the global logger's output writer.
func SetOutputWriter(writer io.Writer) {
	globalLogger.SetOutputWriter(writer)
}

// SetOutputFile sets the global logger's output file.
func SetOutputFile(filename string, logAppend bool, strategy RotationStrategy) error {
	return globalLogger.SetOutputFile(filename, logAppend, strategy)
}

// SetDateFormat sets the date format for the global logger.
func SetDateFormat(dateFormat string) {
	globalLogger.SetDateFormat(dateFormat)
}

// Flush flushes the global logger's write buffer.
func Flush() {
	globalLogger.Flush()
}

// Rotate rotate's the global logger's log file, if possible.
func Rotate() (string, error) {
	return globalLogger.Rotate()
}
