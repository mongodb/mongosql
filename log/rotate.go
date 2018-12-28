package log

import (
	"fmt"
	"os"
	"time"

	"github.com/10gen/sqlproxy/internal/procutil"
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

// RotatedFileName returns the formatted string for a filename with the provided
// base filename, timestamp, and sequence number.
func RotatedFileName(filename, timestamp string, sequenceNumber uint64) string {
	return fmt.Sprintf("%s.%06d.%s", filename, sequenceNumber, timestamp)
}

type rotateFunc func() (string, error)

type rotatingFile struct {
	lastLogTime  time.Time
	logSeqNumber uint64
	filename     string
	mode         int
	strategy     RotationStrategy

	file    *os.File
	lastLog string

	rotateChan    chan struct{}
	errChan       chan error
	dataChan      chan []byte
	writeDoneChan chan struct{}
}

func newRotatingFile(file string, willAppend bool,
	strategy RotationStrategy) (*rotatingFile, error) {

	switch strategy {
	case Rename, Reopen:
		// these are supported
	default:
		return nil, fmt.Errorf("Unsupported log rotation strategy '%s'", strategy)
	}

	// calculate mode for opening file
	mode := os.O_CREATE | os.O_WRONLY
	if willAppend {
		mode = mode | os.O_APPEND
	} else {
		mode = mode | os.O_TRUNC
	}

	rf := &rotatingFile{
		filename:      file,
		mode:          mode,
		strategy:      strategy,
		rotateChan:    make(chan struct{}, 1),
		errChan:       make(chan error),
		dataChan:      make(chan []byte),
		writeDoneChan: make(chan struct{}),
		lastLogTime:   time.Now(),
		logSeqNumber:  0,
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

// The start function kicks off the goroutine that takes care of all disk
// operations (i.e. rotating the log file and flushing the log buffer to disk).
// This function should be called exactly once when the rotatingFile is created.
// We perform all disk operations in this one goroutine because it guarantees
// that there is only one disk operation occurring at any given time. We use
// channels to signal when this goroutine should execute a disk operation and
// to return results back to the signalling goroutine.
func (rf *rotatingFile) start() {
	procutil.PanicSafeGo(func() {
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
					currentTime := time.Now()
					logTimeStamp := currentTime.Format(RotationTimeFormat)

					// Since we are using wall clock to get timestamp for log file, there is a
					// possiblity that we can get same timestamp for two log files or timestamp in
					// the past.  To handle these cases we are adding seqNumber with the log files,
					// which will be incremented whenever either of the two cases mentioned above
					// will be detected.
					if logTimeStamp <= rf.lastLogTime.Format(RotationTimeFormat) {
						rf.logSeqNumber++
					} else {
						rf.lastLogTime = currentTime
					}

					archive := RotatedFileName(rf.filename, logTimeStamp, rf.logSeqNumber)
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

// rotate sends a signal to the disk-operation goroutine to rotate the log file.
// The signalling channel (rf.rotateChan) has a buffer size of one. This
// guarantees that, after our non-blocking send, there is a value waiting in the
// signal channel (if it were an unbuffered channel, a non-blocking send could
// result in no signal being delivered to the disk-operation goroutine if the
// disk-operation goroutine is not waiting to receive on the channel at the same
// time the send attempt is made).
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
