// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package msg

import "testing"

var (
	treadRead = []RRow{
		{
			R: &Tread{},
			B: []byte{0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}
	rreadRead = []RRow{
		{
			R: &Rread{},
			B: []byte{0x0d, 0x00},
		},
	}
)

func TestTreadRead(t *testing.T) {
	readerTest(t, treadRead)
}

func TestRreadRead(t *testing.T) {
	readerTest(t, rreadRead)
}
