// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package msg

import (
	"fmt"
	"io"
)

// Tping is the client pinging the server.
//
//	The ping message has no effect and is used mainly for debug-
//	ging.  Servers should respond immediately to pings.
type Tping struct {
	Tag byte
}

func (m *Tping) Read(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, ErrBufTooSmall
	}
	b[0] = KindTping
	b[1] = m.Tag
	return 2, io.EOF
}

func (m *Tping) Write(b []byte) (int, error) {
	if len(b) < 1 {
		return 0, ErrBufTooSmall
	}
	m.Tag = b[0]
	return 1, nil
}

func (m *Tping) String() string {
	return fmt.Sprintf("Tping tag(%x)", m.Tag)
}

// Rping is the ping response.
type Rping struct {
	Tag byte
}

func (m *Rping) Read(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, ErrBufTooSmall
	}
	b[0] = KindRping
	b[1] = m.Tag
	return 2, io.EOF
}

func (m *Rping) Write(b []byte) (int, error) {
	if len(b) < 1 {
		return 0, ErrBufTooSmall
	}
	m.Tag = b[0]
	return 1, nil
}

func (m *Rping) String() string {
	return fmt.Sprintf("Rping tag(%x)", m.Tag)
}
