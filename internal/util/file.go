package util

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
)

// GetFieldsFromFile fetches the first line from the contents of the file
// at "path"
func GetFieldsFromFile(path string) (fields []string, err error) {
	var fieldFileReader *os.File
	fieldFileReader, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer CheckDeferredFunc(fieldFileReader.Close, &err)

	fieldScanner := bufio.NewScanner(fieldFileReader)
	for fieldScanner.Scan() {
		fields = append(fields, fieldScanner.Text())
	}
	if err = fieldScanner.Err(); err != nil {
		return nil, err
	}
	return fields, nil
}

// ToUniversalPath returns the result of replacing each slash ('/') character
// in "path" with an OS-sepcific separator character. Multiple slashes are
// replaced by multiple separators
func ToUniversalPath(path string) string {
	return filepath.FromSlash(path)
}

// WrappedReadCloser is a wrapper for two - an inner and outer - ReadClosers.
type WrappedReadCloser struct {
	io.ReadCloser
	Inner io.ReadCloser
}

// Close closes both the inner and outer ReadClosers and
// returns an error in performing the operations.
func (wrc *WrappedReadCloser) Close() error {
	outerErr := wrc.ReadCloser.Close()
	innerErr := wrc.Inner.Close()
	if outerErr != nil {
		return outerErr
	}
	return innerErr
}

// WrappedWriteCloser is a wrapper for two - an inner and outer - WriteClosers.
type WrappedWriteCloser struct {
	io.WriteCloser
	Inner io.WriteCloser
}

// Close closes both the inner and outer WriteClosers and
// returns an error in performing the operations.
func (wwc *WrappedWriteCloser) Close() error {
	outerErr := wwc.WriteCloser.Close()
	innerErr := wwc.Inner.Close()
	if outerErr != nil {
		return outerErr
	}
	return innerErr
}
