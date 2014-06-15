// Package temp implements a ReadSeeker backed by memory ([]bytes)
// or a temporary file depending size limit
//
// Copyright 2013 The Agostle Authors. All rights reserved.
// Use of this source code is governed by an Apache 2.0
// license that can be found in the LICENSE file.
package temp

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"time"
)

// ReadSeekCloser is an io.Reader + io.Seeker + io.Closer
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
	Stat() (os.FileInfo, error)
}

type tempBuf struct {
	*bytes.Reader
	stat os.FileInfo
}

//Close implements io.Close (NOP)
func (b *tempBuf) Close() error { //NopCloser
	return nil
}

func (b *tempBuf) Stat() (os.FileInfo, error) {
	return b.stat, nil
}

// NewReadSeeker returns a copy of the r io.Reader which can be Seeken and closed.
func NewReadSeeker(r io.Reader, maxMemory int64) (ReadSeekCloser, error) {
	b := bytes.NewBuffer(nil)
	if maxMemory <= 0 {
		maxMemory = 1 << 20 // 1Mb
	}
	size, err := io.CopyN(b, r, maxMemory+1)
	if err != nil && err != io.EOF {
		return nil, err
	}
	var fi os.FileInfo
	if fh, ok := r.(*os.File); ok {
		fi, err = fh.Stat()
	}
	if size <= maxMemory {
		if fi == nil {
			fi = dummyFileInfo{name: "memory", size: int64(len(b.Bytes())),
				mtime: time.Now()}
		}
		return &tempBuf{Reader: bytes.NewReader(b.Bytes()), stat: fi}, nil
	}
	// too big, write to disk and flush buffer
	file, err := ioutil.TempFile("", "reader-")
	if err != nil {
		return nil, err
	}
	nm := file.Name()
	size, err = io.Copy(file, io.MultiReader(b, r))
	if err != nil {
		file.Close()
		os.Remove(nm)
		return nil, err
	}
	file.Close()
	fh, err := os.Open(nm)
	if err == nil {
		if e := os.Remove(nm); e != nil && !os.IsNotExist(e) {
			deleteLater(nm)
		}
	}
	if fi == nil {
		fi, err = fh.Stat()
	}
	return &tempFile{File: fh, stat: fi}, err
}

type tempFile struct {
	*os.File
	stat os.FileInfo
}

func (f *tempFile) Close() error {
	nm := f.Name()
	f.File.Close()
	return os.Remove(nm)
}

func (f *tempFile) Stat() (os.FileInfo, error) {
	return f.File.Stat()
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
