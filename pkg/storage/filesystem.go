// Copyright (c) 2021 UMI
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package storage

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

var (
	ErrNotExist   = fs.ErrNotExist
	ErrExist      = fs.ErrExist
	ErrPermission = fs.ErrPermission
	ErrIsDir      = errors.New("not a file")
	ErrNotDir     = errors.New("not a directory")
	ErrSize       = errors.New("file has wrong size")
)

type iFS interface {
	Stat(string) (fs.FileInfo, error)
	MkdirAll(string, fs.FileMode) error
	Create(string) (IFile, error)
	OpenFile(string) (IFile, error)
}

type IFile interface {
	io.WriteCloser
	io.WriterAt
	io.ReaderAt
	Stat() (fs.FileInfo, error)
}

type FSx struct{}

func NewFSx() *FSx {
	return &FSx{}
}

func (FSx) Stat(name string) (fi fs.FileInfo, err error) {
	if fi, err = os.Stat(name); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return fi, nil
}

func (FSx) MkdirAll(path string, perm fs.FileMode) error {
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (FSx) Create(name string) (file IFile, err error) {
	if file, err = os.Create(name); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return file, nil
}

func (FSx) OpenFile(name string) (IFile, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0o644)
	if err != nil {
		err = fmt.Errorf("%w", err)
	}

	return file, err
}

func CheckOrCreateDir(fsx iFS, path string) error {
	fi, err := fsx.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return nil
		}

		return ErrNotDir
	}

	if errors.Is(err, fs.ErrNotExist) {
		if err = fsx.MkdirAll(path, 0o755); err != nil {
			err = fmt.Errorf("%w", err)
		}
	}

	return err
}

func OpenOrCreateFile(fsx iFS, name string, size int) (file IFile, err error) {
	fileInfo, err := fsx.Stat(name)

	if err == nil {
		if fileInfo.IsDir() {
			return nil, ErrIsDir
		}

		if size != 0 && fileInfo.Size() != int64(size) {
			return nil, ErrSize
		}

		if file, err = fsx.OpenFile(name); err != nil {
			err = fmt.Errorf("%w", err)
		}

		return file, err
	}

	if errors.Is(err, fs.ErrNotExist) {
		return CreateAndFillFile(fsx, name, size)
	}

	return nil, fmt.Errorf("%w", err)
}

func CreateAndFillFile(fsx iFS, name string, size int) (file IFile, err error) {
	if file, err = fsx.Create(name); err != nil {
		err = fmt.Errorf("create error: %w", err)

		return nil, err
	}

	bufLength := 1024 * 1024 // 1Mb
	buf := make([]byte, bufLength)

	// Заполняем нулями
	for size > 0 {
		if bufLength > size {
			bufLength = size
		}

		if _, err = file.Write(buf[0:bufLength]); err != nil {
			err = fmt.Errorf("write error: %w", err)

			return nil, err
		}

		size -= bufLength
	}

	return file, nil
}
