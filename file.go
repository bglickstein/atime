package atime

import (
	"io"
	"os"
	"time"
)

// Satisfies io.ReadCloser. Stats the file before opening it and
// restores the mtime and atime when closing it.
type FileReadCloser struct {
	f            *os.File
	mtime, atime time.Time
}

func NewFileReadCloser(path string) (*FileReadCloser, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &FileReadCloser{
		f:     f,
		mtime: fi.ModTime(),
		atime: Get(fi),
	}, nil
}

func (a FileReadCloser) Read(p []byte) (int, error) {
	return a.f.Read(p)
}

func (a FileReadCloser) Close() error {
	path := a.f.Name()
	err := a.f.Close()
	if err != nil {
		return err
	}
	return os.Chtimes(path, a.atime, a.mtime)
}

func WithTimesRestored(path string, fn func(io.Reader) error) error {
	r, err := NewFileReadCloser(path)
	if err != nil {
		return err
	}
	defer r.Close()
	return fn(r)
}
