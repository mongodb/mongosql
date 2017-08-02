package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	Rename = "rename"
	Reopen = "reopen"
)

type rotateFunc func() (string, error)

type rotatingfile struct {
	filename string
	mode     int
	strategy string

	file    *os.File
	lastLog string

	rotateChan chan struct{}
	errChan    chan error

	reader *io.PipeReader
	writer *io.PipeWriter

	memBuf []byte
}

func newRotatingFile(filename string, append bool, strategy string) (io.Writer, rotateFunc, error) {

	switch strategy {
	case Rename, Reopen:
		// these are supported
	default:
		return nil, nil, fmt.Errorf("Unsupported log rotation strategy '%s'", strategy)
	}

	// calculate mode for opening file
	mode := os.O_CREATE | os.O_WRONLY
	if append {
		mode = mode | os.O_APPEND
	} else {
		mode = mode | os.O_TRUNC
	}

	// create io pipeline
	pr, pw := io.Pipe()

	rf := &rotatingfile{
		filename:   filename,
		mode:       mode,
		strategy:   strategy,
		rotateChan: make(chan struct{}, 1),
		errChan:    make(chan error),
		reader:     pr,
		writer:     pw,
		memBuf:     make([]byte, 4096),
	}

	w, err := rf.open()
	if err == nil {
		rf.start()
	}
	return w, rf.rotate, err
}

func (rf *rotatingfile) open() (io.Writer, error) {
	// open log file
	file, err := os.OpenFile(rf.filename, rf.mode, 0666)
	if err != nil {
		return nil, err
	}
	rf.file = file

	return rf.writer, nil
}

func (f *rotatingfile) start() {
	data := make(chan []byte)
	var chunkSize int64 = 32

	go func() {
		for {
			buf := make([]byte, chunkSize)
			n, err := f.reader.Read(buf)
			if err == nil {
				data <- buf[:n]
			}
		}
	}()

	go func() {
		tickChan := time.NewTicker(time.Millisecond * 400).C
		for {
			select {
			case <-f.rotateChan:
				var err error

				if len(f.memBuf) != 0 {
					f.file.Write(f.memBuf)
					f.memBuf = f.memBuf[:0]
				}

				if f.strategy == Rename {
					// rename current log file to the archived log file
					now := time.Now().Format(time.RFC3339Nano)
					archive := fmt.Sprintf("%s.%s", f.filename, now)
					err = os.Rename(f.filename, archive)
					if err != nil {
						// if rename fails, we will just keep logging to the same file
						f.errChan <- fmt.Errorf("Log rotation failed: %v", err)
						break
					}
					f.lastLog = archive
				}

				// close the log file
				err = f.file.Close()
				if err != nil {
					panic(fmt.Errorf("Log rotation failed: %v", err))
				}

				// open the new log file
				_, err = f.open()
				if err != nil {
					panic(fmt.Errorf("Log rotation failed: %v", err))
				}

				f.errChan <- nil
			case buf := <-data:
				f.memBuf = append(f.memBuf, buf...)
			case <-tickChan:
				f.file.Write(f.memBuf)
				f.memBuf = f.memBuf[:0]
			}
		}
	}()
}

func (f *rotatingfile) rotate() (string, error) {
	select {
	case f.rotateChan <- struct{}{}:
		// signal a rotation
	default:
		return "", fmt.Errorf("Log rotation already in progress")
	}

	// wait for rotation to complete
	err := <-f.errChan
	return f.lastLog, err
}
