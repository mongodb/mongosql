package log

import (
	"fmt"
	"os"
	"time"

	"github.com/10gen/sqlproxy/internal/util"
)

// RotationStrategy is an enum representing mongosqld's supported log rotation
// strategies.
type RotationStrategy string

// Possible values for RotationStrategy
const (
	Rename RotationStrategy = "rename"
	Reopen RotationStrategy = "reopen"
)

// RotationTimeFormat is the timestamp format used in the filenames of rotated
// log files. It is equivalent time.RFC3339Nano with "_" instead of ":"
// (because you cannot have colons in Windows filenames).
const RotationTimeFormat = "2006-01-02T15_04_05.999999999Z07_00"

type rotateFunc func() (string, error)

type rotatingFile struct {
	filename string
	mode     int
	strategy RotationStrategy

	file    *os.File
	lastLog string

	rotateChan    chan struct{}
	errChan       chan error
	dataChan      chan []byte
	writeDoneChan chan struct{}
}

func newRotatingFile(filename string, append bool, strategy RotationStrategy) (*rotatingFile, error) {

	switch strategy {
	case Rename, Reopen:
		// these are supported
	default:
		return nil, fmt.Errorf("Unsupported log rotation strategy '%s'", strategy)
	}

	// calculate mode for opening file
	mode := os.O_CREATE | os.O_WRONLY
	if append {
		mode = mode | os.O_APPEND
	} else {
		mode = mode | os.O_TRUNC
	}

	rf := &rotatingFile{
		filename:      filename,
		mode:          mode,
		strategy:      strategy,
		rotateChan:    make(chan struct{}, 1),
		errChan:       make(chan error),
		dataChan:      make(chan []byte),
		writeDoneChan: make(chan struct{}),
	}

	err := rf.open()
	if err == nil {
		rf.start()
	}
	return rf, err
}

func (rf *rotatingFile) Write(b []byte) (int, error) {
	rf.dataChan <- b
	<-rf.writeDoneChan
	return len(b), nil
}

func (rf *rotatingFile) open() error {
	// open log file
	file, err := os.OpenFile(rf.filename, rf.mode, 0666)
	if err != nil {
		return err
	}
	rf.file = file

	return nil
}

func (rf *rotatingFile) start() {
	util.PanicSafeGo(func() {
		for {
			select {
			case <-rf.rotateChan:
				var err error

				// close the log file
				err = rf.file.Close()
				if err != nil {
					panic(fmt.Errorf("Log rotation failed: %v", err))
				}

				if rf.strategy == Rename {
					// rename current log file to the archived log file
					now := time.Now().Format(RotationTimeFormat)
					archive := fmt.Sprintf("%s.%s", rf.filename, now)
					err = os.Rename(rf.filename, archive)
					if err != nil {
						panic(fmt.Errorf("Log rotation failed: %v", err))
					}
					rf.lastLog = archive
				}

				// open the new log file
				err = rf.open()
				if err != nil {
					panic(fmt.Errorf("Log rotation failed: %v", err))
				}

				rf.errChan <- nil
			case b := <-rf.dataChan:
				_, err := rf.file.Write(b)
				if err != nil {
					panic(fmt.Errorf("Failed writing log: %v", err))
				}
				rf.writeDoneChan <- struct{}{}
			}
		}
	}, func(err interface{}) {
		loggingComponentLogger().Fatalf(Always, "%v", err)
		panic(err)
	})
}

func (rf *rotatingFile) rotate() (string, error) {
	select {
	case rf.rotateChan <- struct{}{}:
		// signal a rotation
	default:
		return "", fmt.Errorf("Log rotation already in progress")
	}

	// wait for rotation to complete
	err := <-rf.errChan
	return rf.lastLog, err
}
