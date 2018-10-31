package log

import (
	"os"
	"syscall"
)

// Windows-specific variables needed for redirecting stdout and stderr.
var (
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
	procSetStdHandle = kernel32.MustFindProc("SetStdHandle")
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
	err = setStdHandle(syscall.STD_OUTPUT_HANDLE, syscall.Handle(file.Fd()))
	if err != nil {
		return err
	}

	// Redirect standard error to a file.
	err = setStdHandle(syscall.STD_ERROR_HANDLE, syscall.Handle(file.Fd()))
	if err != nil {
		return err
	}

	// Calling SetStdHandle does not affect prior references to stderr or
	// stdout. Set the os handles to stdout and stderr explicitly.
	os.Stdout = file
	os.Stderr = file
	rf.file = file

	return nil
}

func setStdHandle(stdhandle int32, handle syscall.Handle) error {
	r0, _, e1 := syscall.Syscall(procSetStdHandle.Addr(), 2, uintptr(stdhandle), uintptr(handle), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}
