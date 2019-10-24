// +build linux

package log

import (
	"os"
	"syscall"
)

// The open function opens the file at rf.filename with mode rf.mode, and
// sets it as the current output file. Because this is a disk operation, it
// should only ever be called from the asynchronous disk goroutine created
// in the start() function.
func (rf *rotatingFile) open() error {
	file, err := os.OpenFile(rf.filename, rf.mode, 0666)
	if err != nil {
		return err
	}

	// Redirect standard output to a file.
	err = syscall.Dup3(int(file.Fd()), syscall.Stdout, 0)
	if err != nil {
		return err
	}

	// Redirect standard error to a file.
	err = syscall.Dup3(int(file.Fd()), syscall.Stderr, 0)
	if err != nil {
		return err
	}
	rf.file = file

	return nil
}
