/*
Copied from camlistore.org/pkg/blobstore/localdisk/receive.go

Copyright 2011 Google Inc.

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
	"fmt"
	"io"
	"os"
)

// LinkAlreadyExists checks the error and returns whether this is about
// a link already exists or not
func LinkAlreadyExists(err error) bool {
	if os.IsExist(err) {
		return true
	}
	if le, ok := err.(*os.LinkError); ok && os.IsExist(le.Err) {
		return true
	}
	return false
}

// Used by Windows (receive_windows.go) and when a posix filesystem doesn't
// support a link operation (e.g. Linux with an exfat external USB disk).
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file %q: %s", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file %q: %s", dst, err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("error copying from %q to %q: %s", src, dst, err)
	}
	return nil
}
