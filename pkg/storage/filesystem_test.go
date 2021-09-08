package storage_test

import (
	"errors"
	"io/fs"
	"testing"
	"time"

	. "gitlab.com/umitop/umid/pkg/storage"
)

type MockFile struct {
	writeError error
	readError  error
	isDir      bool
	modTime    time.Time
	data       []byte
}

func (mockFile *MockFile) Name() string {
	return ""
}

func (mockFile *MockFile) Size() int64 {
	return int64(len(mockFile.data))
}

func (mockFile *MockFile) Mode() fs.FileMode {
	return 0755
}

func (mockFile *MockFile) ModTime() time.Time {
	return mockFile.modTime
}

func (mockFile *MockFile) IsDir() bool {
	return mockFile.isDir
}

func (mockFile *MockFile) Sys() interface{} {
	return nil
}

func (mockFile *MockFile) Close() error {
	return nil
}

func (mockFile *MockFile) Write(data []byte) (int, error) {
	if mockFile.writeError != nil {
		return 0, mockFile.writeError
	}

	mockFile.data = append(mockFile.data, data...)

	return len(data), nil
}

func (mockFile *MockFile) WriteAt(data []byte, offset int64) (int, error) {
	copy(mockFile.data[offset:], data)

	return 0, nil
}

func (mockFile *MockFile) ReadAt(data []byte, offset int64) (int, error) {
	copy(data, mockFile.data[offset:])

	return 0, nil
}

func (mockFile *MockFile) Stat() (fs.FileInfo, error) {
	return mockFile, nil
}

type MockFs struct {
	files        map[string]*MockFile
	createErrors map[string]error
	writeErrors  map[string]error
}

func NewMockFs() *MockFs {
	return &MockFs{
		files:        make(map[string]*MockFile),
		createErrors: make(map[string]error),
		writeErrors:  make(map[string]error),
	}
}

func (mock *MockFs) Stat(name string) (fs.FileInfo, error) {
	mockFile, ok := mock.files[name]

	if !ok {
		return nil, fs.ErrNotExist
	}

	return mockFile, nil
}

func (mock *MockFs) MkdirAll(path string, perm fs.FileMode) error {
	if _, ok := mock.files[path]; ok {
		return ErrExist
	}

	if err, ok := mock.createErrors[path]; ok {
		return err
	}

	mock.files[path] = &MockFile{
		isDir:   true,
		modTime: time.Now(),
	}

	return nil
}

func (mock *MockFs) Create(name string) (IFile, error) {
	if _, ok := mock.files[name]; ok {
		return nil, ErrExist
	}

	if err, ok := mock.createErrors[name]; ok {
		return nil, err
	}

	file := &MockFile{
		isDir:   false,
		modTime: time.Now(),
	}

	if err, ok := mock.writeErrors[name]; ok {
		file.writeError = err
	}

	mock.files[name] = file

	return file, nil
}

func (mock *MockFs) OpenFile(name string) (IFile, error) {
	if file, ok := mock.files[name]; ok {
		return file, nil
	}

	return nil, ErrNotExist
}

// Функция CheckOrCreateDir должна вернуть nil, если файл уже существет и является директорией.
func TestFS001(t *testing.T) {
	t.Parallel()

	dir := "/tmp"

	mockFs := NewMockFs()
	mockFs.files[dir] = &MockFile{
		isDir: true,
	}

	err := CheckOrCreateDir(mockFs, dir)

	if err != nil {
		t.Errorf("ожидаем 'nil', получили '%v'", err)
	}
}

// Функция CheckOrCreateDir должна вернуть ошибку, если файл уже существет и он не директорией.
func TestFS002(t *testing.T) {
	t.Parallel()

	exp := ErrNotDir
	dir := "/tmp"

	mockFs := NewMockFs()
	mockFs.files[dir] = &MockFile{
		isDir: false,
	}

	err := CheckOrCreateDir(mockFs, dir)

	if !errors.Is(err, exp) {
		t.Errorf("ожидаем '%v', получили '%v'", exp, err)
	}
}

// Функция CheckOrCreateDir должна вернуть ошибку если дериктории не сушествует, но не хватает прав создать ее.
func TestFS003(t *testing.T) {
	t.Parallel()

	exp := ErrPermission
	dir := "/tmp"

	mockFs := NewMockFs()
	mockFs.createErrors[dir] = exp

	err := CheckOrCreateDir(mockFs, dir)

	if !errors.Is(err, exp) {
		t.Errorf("ожидаем '%v', получили '%v'", exp, err)
	}
}

// Функция CreateAndFillFile должна вернуть ошибку если файл уже существует
func TestFS004(t *testing.T) {
	t.Parallel()

	exp := ErrExist
	name := "/tmp/file"

	mockFs := NewMockFs()
	mockFs.files[name] = nil

	_, err := CreateAndFillFile(mockFs, name, 1024)

	if !errors.Is(err, exp) {
		t.Errorf("ожидаем '%v', получили '%v'", exp, err)
	}
}

// Функция CreateAndFillFile должна вернуть ошибку если файл не сущестуе, но не хватает прав его создать
func TestFS005(t *testing.T) {
	t.Parallel()

	exp := ErrPermission
	name := "/tmp/file"
	size := 0

	mockFs := NewMockFs()
	mockFs.createErrors[name] = exp

	_, err := CreateAndFillFile(mockFs, name, size)

	if !errors.Is(err, exp) {
		t.Errorf("ожидаем '%v', получили '%v'", exp, err)
	}
}

// Функция CreateAndFillFile должна вернуть успешно созданный файл правильного размера если все ОК.
func TestFS006(t *testing.T) {
	t.Parallel()

	name := "/tmp/file"
	size := 1024

	mockFs := NewMockFs()

	file, err := CreateAndFillFile(mockFs, name, size)

	if err != nil {
		t.Errorf("ожидаем 'nil', получили '%v'", err)
	}

	if fi, _ := file.Stat(); fi.Size() != int64(size) {
		t.Errorf("ожидаем '%d', получили '%v'", size, fi.Size())
	}
}

// Функция CreateAndFillFile должна вернуть ошибку если файл успешно создан, но не удалось его заполнить нулями.
func TestFS007(t *testing.T) {
	t.Parallel()

	exp := ErrPermission
	name := "/tmp/file"
	size := 1024

	mockFs := NewMockFs()
	mockFs.writeErrors[name] = ErrPermission

	_, err := CreateAndFillFile(mockFs, name, size)

	if !errors.Is(err, exp) {
		t.Errorf("ожидаем '%v', получили '%v'", exp, err)
	}
}
