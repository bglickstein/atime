package atime

import (
	"io"
	"os"
	"time"

	orig "github.com/djherbis/atime"
)

// File satisfies io.Reader, ReaderAt, and Seeker,
// delegating each method to the *os.File it contains.
// It also implements io.Closer.
//
// When a File is created with Open, it is statted
// and its mtime and atime are recorded.
// When a File is closed, its mtime and atime are restored.
type File struct {
	io.ReadCloser
	io.ReaderAt
	io.Seeker

	// F is the *os.File to which this object's methods are delegated.
	F *os.File

	mtime, atime time.Time
}

// Open opens a new File for reading.
// The caller should close the File when finished.
// Closing will restore the file's mtime and atime.
func Open(path string) (*File, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &File{
		F:     f,
		mtime: fi.ModTime(),
		atime: orig.Get(fi),
	}, nil
}

// Read implements io.Reader.
func (a File) Read(p []byte) (int, error) {
	return a.F.Read(p)
}

// ReadAt implements io.ReaderAt.
func (a File) ReadAt(p []byte, off int64) (int, error) {
	return a.F.ReadAt(p, off)
}

// Seek implements io.Seeker.
func (a File) Seek(offset int64, whence int) (int64, error) {
	return a.F.Seek(offset, whence)
}

// Close closes the file,
// restoring the mtime and atime to what they were when the file was first opened.
func (a File) Close() error {
	path := a.F.Name()
	err := a.F.Close()
	if err != nil {
		return err
	}
	return os.Chtimes(path, a.atime, a.mtime)
}

// WithTimesRestored opens the named file, passes it to a callback,
// and closes it afterward, restoring its atime and mtime.
func WithTimesRestored(path string, fn func(io.ReadSeeker) error) error {
	r, err := Open(path)
	if err != nil {
		return err
	}
	defer r.Close()
	return fn(r)
}
