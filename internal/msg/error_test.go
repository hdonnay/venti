// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package msg

import "testing"

var (
	rerrorRead  = []RRow{}
	rerrorWrite = []WRow{
		{
			A: &Rerror{},
			B: &Rerror{},
		},
	}
)

func TestRerrorRead(t *testing.T) {
	readerTest(t, rerrorRead)
}

func TestRerrorWrite(t *testing.T) {
	writerTest(t, rerrorWrite)
}
