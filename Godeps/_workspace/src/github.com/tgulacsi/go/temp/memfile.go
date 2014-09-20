/*
Copyright 2013 the Camlistore authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package temp

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"time"
)

var MaxInMemorySlurp = 4 << 20 // 4MB.  *shrug*.

type ReaderAt interface {
	ReadAt(p []byte, off int64) (n int, err error)
}

type Stater interface {
	Stat() (os.FileInfo, error)
}

// ReadSeekCloser is an io.Reader + ReaderAt + io.Seeker + io.Closer + Stater
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	ReaderAt
	io.Closer
	Stat() (os.FileInfo, error)
}

// MakeReadSeekCloser makes an io.ReadSeeker + io.Closer by reading the whole reader
// If the given Reader is a Closer, too, than that Close will be called
func MakeReadSeekCloser(blobRef string, r io.Reader) (ReadSeekCloser, error) {
	if rc, ok := r.(io.Closer); ok {
		defer rc.Close()
	}
	ms := NewMemorySlurper(blobRef)
	n, err := io.Copy(ms, r)
	if err != nil {
		return nil, err
	}

	if fh, ok := r.(*os.File); ok {
		ms.stat, err = fh.Stat()
	}
	if ms.stat == nil {
		if ms.file == nil {
			ms.stat = dummyFileInfo{name: "memory", size: n, mtime: time.Now()}
		} else {
			ms.stat, err = ms.file.Stat()
		}
	}
	return ms, err
}

// NewReadSeeker is a convenience function of MakeReadSeekCloser.
func NewReadSeeker(r io.Reader) (ReadSeekCloser, error) {
	return MakeReadSeekCloser("", r)
}

// memorySlurper slurps up a blob to memory (or spilling to disk if
// over MaxInMemorySlurp) and deletes the file on Close
type memorySlurper struct {
	maxInMemorySlurp int
	blobRef          string // only used for tempfile's prefix
	buf              *bytes.Buffer
	mem              *bytes.Reader
	file             *os.File // nil until allocated
	reading          bool     // transitions at most once from false -> true
	stat             os.FileInfo
}

func NewMemorySlurper(blobRef string) *memorySlurper {
	return &memorySlurper{
		blobRef: blobRef,
		buf:     new(bytes.Buffer),
	}
}

func (ms *memorySlurper) prepareRead() {
	if ms.reading {
		return
	}
	ms.reading = true
	if ms.file != nil {
		ms.file.Seek(0, 0)
	} else {
		ms.mem = bytes.NewReader(ms.buf.Bytes())
		ms.buf = nil
	}
}

func (ms *memorySlurper) Read(p []byte) (n int, err error) {
	ms.prepareRead()
	if ms.file != nil {
		return ms.file.Read(p)
	}
	return ms.mem.Read(p)
}

func (ms *memorySlurper) ReadAt(p []byte, off int64) (n int, err error) {
	ms.prepareRead()
	if ms.file != nil {
		return ms.file.ReadAt(p, off)
	}
	return ms.mem.ReadAt(p, off)
}

func (ms *memorySlurper) Seek(offset int64, whence int) (int64, error) {
	if !ms.reading {
		ms.reading = true
		if ms.file == nil {
			ms.mem = bytes.NewReader(ms.buf.Bytes())
			ms.buf = nil
		}
	}
	if ms.file != nil {
		return ms.file.Seek(offset, whence)
	}
	return ms.mem.Seek(offset, whence)
}

func (ms *memorySlurper) ReadFrom(r io.Reader) (n int64, err error) {
	if ms.reading {
		panic("write after read")
	}
	if ms.maxInMemorySlurp <= 0 {
		ms.maxInMemorySlurp = MaxInMemorySlurp
	}
	var size int64
	if fh, ok := r.(*os.File); ok {
		ms.stat, err = fh.Stat()
		size = ms.stat.Size()
	}
	if ms.file == nil && size > 0 && size < int64(ms.maxInMemorySlurp) {
		return io.Copy(ms.buf, r)
	}
	if ms.file == nil {
		ms.file, err = ioutil.TempFile("", ms.blobRef)
		if err != nil {
			return 0, err
		}
	}
	return io.Copy(ms.file, r)
}

func (ms *memorySlurper) Write(p []byte) (n int, err error) {
	if ms.reading {
		panic("write after read")
	}
	if ms.file != nil {
		n, err = ms.file.Write(p)
		return
	}

	if ms.maxInMemorySlurp <= 0 {
		ms.maxInMemorySlurp = MaxInMemorySlurp
	}
	if ms.buf.Len()+len(p) > ms.maxInMemorySlurp {
		ms.file, err = ioutil.TempFile("", ms.blobRef)
		if err != nil {
			return
		}
		_, err = io.Copy(ms.file, ms.buf)
		if err != nil {
			return
		}
		ms.buf = nil
		n, err = ms.file.Write(p)
		return
	}

	return ms.buf.Write(p)
}

func (ms *memorySlurper) Cleanup() error {
	if ms.file != nil {
		return os.Remove(ms.file.Name())
	}
	return nil
}

func (ms *memorySlurper) Close() error {
	return ms.Cleanup()
}

func (ms *memorySlurper) Stat() (os.FileInfo, error) {
	return ms.stat, nil
}

type dummyFileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	mtime time.Time
	isDir bool
}

func (dfi dummyFileInfo) Name() string {
	return dfi.name
}
func (dfi dummyFileInfo) Size() int64 {
	return dfi.size
}
func (dfi dummyFileInfo) Mode() os.FileMode {
	return dfi.mode
}
func (dfi dummyFileInfo) ModTime() time.Time {
	return dfi.mtime
}
func (dfi dummyFileInfo) IsDir() bool {
	return dfi.isDir
}
func (dfi dummyFileInfo) Sys() interface{} {
	return nil
}
