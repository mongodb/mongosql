// +build aix darwin dragonfly freebsd js,wasm linux nacl netbsd openbsd solaris

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
	err = syscall.Dup2(int(file.Fd()), syscall.Stdout)
	if err != nil {
		return err
	}

	// Redirect standard error to a file.
	err = syscall.Dup2(int(file.Fd()), syscall.Stderr)
	if err != nil {
		return err
	}
	rf.file = file

	return nil
}
