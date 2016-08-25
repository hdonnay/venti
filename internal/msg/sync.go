// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package msg

import (
	"fmt"
	"io"
)

// Tsync is VtTsync.
type Tsync struct {
	Tag byte
}

func (m *Tsync) Read(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, io.ErrShortBuffer
	}
	b[0] = KindTsync
	b[1] = m.Tag
	return 2, io.EOF
}

func (m *Tsync) Write(b []byte) (int, error) {
	if len(b) < 1 {
		return 0, ErrBufTooSmall
	}
	m.Tag = b[0]
	return 1, nil
}

func (m *Tsync) String() string {
	return fmt.Sprintf("Tsync tag(%x)", m.Tag)
}

// Rsync is VtRsync.
type Rsync struct {
	Tag byte
}

func (m *Rsync) Read(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, io.ErrShortBuffer
	}
	b[0] = KindRsync
	b[1] = m.Tag
	return 2, io.EOF
}

func (m *Rsync) Write(b []byte) (int, error) {
	if len(b) < 1 {
		return 0, ErrBufTooSmall
	}
	m.Tag = b[0]
	return 1, nil
}

func (m *Rsync) String() string {
	return fmt.Sprintf("Rsync tag(%x)", m.Tag)
}
