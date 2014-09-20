/*
  Copyright 2013 Tamás Gulácsi

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

/*
Package temp contains some temporary file -related functions
*/
package temp

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// ReaderToFile copies the reader to a temp file and returns its name or error
func ReaderToFile(r io.Reader, prefix, suffix string) (filename string, err error) {
	dfh, e := ioutil.TempFile("", "agostle-"+BaseName(prefix)+"-")
	if e != nil {
		err = e
		return
	}
	if suffix != "" {
		defer func() {
			if err == nil {
				nfn := filename + suffix
				err = os.Rename(filename, nfn)
				filename = nfn
			}
		}()
	}
	if sfh, ok := r.(*os.File); ok {
		filename = dfh.Name()
		dfh.Close()
		os.Remove(filename)
		err = LinkOrCopy(sfh.Name(), filename)
		return
	}
	if _, err = io.Copy(dfh, r); err == nil {
		filename = dfh.Name()
	}
	dfh.Close()
	return
}

// BaseName returns the last part of the filename - both POSIX and Windows meaning
func BaseName(fileName string) string {
	if fileName == "" {
		return ""
	}
	i := strings.LastIndexAny(fileName, "/\\")
	if i >= 0 {
		fileName = fileName[i+1:]
	}
	return fileName
}
