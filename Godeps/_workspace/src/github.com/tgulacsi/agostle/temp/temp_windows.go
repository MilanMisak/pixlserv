// +build windows

// Copyright 2013 The Agostle Authors. All rights reserved.
// Use of this source code is governed by an Apache 2.0
// license that can be found in the LICENSE file.

package temp

import (
	"os"
	"sync"
	"time"
)

var tbd []string
var tbdMtx = new(sync.Mutex)

func deleteLater(nm string) {
	tbdMtx.Lock()
	defer tbdMtx.Unlock()

	if tbd == nil {
		tbd = []string{nm}
		go func() {
			for _ = range time.Tick(5 * time.Minute) {
				tbdMtx.Lock()
				var e error
				for i := len(tbd) - 1; i >= 0; i-- {
					if e = os.Remove(tbd[i]); e == nil {
						tbd = tbd[:i]
					}
				}
				tbdMtx.Unlock()
			}
		}()
		return
	}
	for _, x := range tbd {
		if x == nm {
			return
		}
	}
	tbd = append(tbd, nm)
}
