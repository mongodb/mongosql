// Package log provides a utility to log timestamped messages to an io.Writer.
package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
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
	defaultTimeFormat        = "2006-01-02T15:04:05.000-0700"
	flushInterval            = 400 * time.Millisecond
	bufferSizeFlushThreshold = 4096
	bufferSizeLimit          = 8192
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

//OS will not change mid-run, so we set our logging line endings once, on package init
func getNewLine() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

var (
	level = map[int]string{
		Always:    "I", // Informational
		Info:      "I",
		DebugLow:  "D", // Debug
		DebugHigh: "D",
	}
	logNewLine = getNewLine()
)

type Logger struct {
	buffer     *writeBuffer
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
	lg.buffer.setDateFormat(dateFormat)
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
	lg.buffer.setWriter(writer)
	lg.rotateFunc = noRotateFunc
}

func (lg *Logger) SetOutputFile(filename string, logAppend bool, strategy RotationStrategy) error {
	w, err := newRotatingFile(filename, logAppend, strategy)
	if err == nil {
		lg.buffer.setWriter(w)
		lg.rotateFunc = w.rotate
	}
	return err
}

func (lg *Logger) Flush() {
	lg.buffer.requestFlush(true)
}

func (lg *Logger) Rotate() (string, error) {
	return lg.rotateFunc()
}

func (lg *Logger) writelog(minVerbosity int, msg string) {
	if minVerbosity < 0 {
		panic("cannot set a minimum log verbosity that is less than 0")
	}

	if minVerbosity <= lg.verbosity {
		lg.buffer.writeMessage(msg)
	}
}

func noRotateFunc() (string, error) {
	return "", fmt.Errorf("cannot rotate logs without log path: use --logPath or in a config " +
		"file at 'systemLog.path'")
}

func NewLogger(verbosity VerbosityLevel) *Logger {
	lg := &Logger{
		buffer:     newWriteBuffer(os.Stderr, bufferSizeFlushThreshold, bufferSizeLimit),
		component:  defaultComponent,
		rotateFunc: noRotateFunc,
	}

	lg.SetVerbosity(verbosity)

	return lg
}

func NewComponentLogger(component string, logger Logger) *Logger {
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

func newWriteBuffer(writer io.Writer, threshold, limit int) *writeBuffer {
	flushChan := make(chan chan struct{}, 1)
	w := &writeBuffer{
		threshold: threshold,
		limit:     limit,
		writer:    writer,
		flushChan: flushChan,
		format:    defaultTimeFormat,
	}

	go func() {
		for {
			select {
			case <-time.After(flushInterval):
				w.flush()
			case done := <-flushChan:
				w.flush()
				done <- struct{}{}
			}
		}
	}()

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
	msgBytes := []byte(fmt.Sprintf("%v %v%s", time.Now().Format(w.format), str, logNewLine))
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

func SetOutputFile(filename string, logAppend bool, strategy RotationStrategy) error {
	return globalLogger.SetOutputFile(filename, logAppend, strategy)
}

func SetDateFormat(dateFormat string) {
	globalLogger.SetDateFormat(dateFormat)
}

func Flush() {
	globalLogger.Flush()
}

func Rotate() (string, error) {
	return globalLogger.Rotate()
}
