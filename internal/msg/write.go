// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package msg

import (
	"fmt"
	"io"
)

// Twrite is VtTwrite.
type Twrite struct {
	Tag  byte
	Type byte
	Pad  [3]byte
}

func (m *Twrite) Read(b []byte) (int, error) {
	if len(b) < 6 {
		return 0, io.ErrShortBuffer
	}
	b[0] = KindTwrite
	b[1] = m.Tag
	b[2] = m.Type
	b[3] = 0x00
	b[4] = 0x00
	b[5] = 0x00
	return 6, io.EOF
}

func (m *Twrite) Write(b []byte) (int, error) {
	if len(b) < 5 {
		return 0, io.ErrShortWrite
	}
	m.Tag = b[0]
	m.Type = b[1]
	copy(m.Pad[:], b[2:])
	return 5, nil
}

func (m *Twrite) String() string {
	return fmt.Sprintf("Twrite tag(%x) type(%x)", m.Tag, m.Type)
}

// Rwrite is VtRwrite.
type Rwrite struct {
	Tag   byte
	Score []byte
}

func (m *Rwrite) Read(b []byte) (int, error) {
	sz := 2 + len(m.Score)
	if len(b) < sz {
		return 0, io.ErrShortBuffer
	}
	b[0] = KindRwrite
	b[1] = m.Tag
	copy(b[2:], m.Score)
	return sz, io.EOF
}

func (m *Rwrite) Write(b []byte) (int, error) {
	if len(b) < 21 {
		return 0, io.ErrShortWrite
	}
	m.Tag = b[0]
	n := copy(m.Score, b[1:])
	return n + 1, nil
}

func (m *Rwrite) String() string {
	return fmt.Sprintf("Rwrite tag(%x) score(%s)", m.Tag, fmtScore(m.Score))
}
