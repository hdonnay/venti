// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package msg

import (
	"bytes"
	"fmt"
	"io"

	"github.com/hdonnay/venti/internal/pack"
)

// Tread is VtTread.
type Tread struct {
	Tag   byte
	Score []byte
	Type  byte
	Pad   byte
	Count uint32
}

func (m *Tread) Read(b []byte) (int, error) {
	sz := 8 + len(m.Score)
	if len(b) < sz {
		return 0, io.ErrShortBuffer
	}
	b[0] = KindTread
	b[1] = m.Tag
	copy(b[2:], m.Score)
	b[sz-6] = m.Type
	b[sz-5] = m.Pad
	pack.Uint32(b[sz-4:], m.Count)
	return sz, io.EOF
}

func (m *Tread) Write(b []byte) (int, error) {
	m.Tag = b[0]
	m.Score = make([]byte, len(b)-7)
	copy(m.Score, b[1:len(b)-6])
	m.Type = b[len(b)-6]
	m.Pad = b[len(b)-5]
	_, m.Count = pack.UnUint32(b[len(b)-4:])
	return len(b), nil
}

func (m *Tread) String() string {
	return fmt.Sprintf("Tread tag(%x) score(%s) type(%x) pad(%x) ct(%d)",
		m.Tag, fmtScore(m.Score), m.Type, m.Pad, m.Count)
}

// Rread is VtRread.
type Rread struct {
	Tag  byte
	Data io.Reader
	hdr  bool
}

func (m *Rread) Read(b []byte) (int, error) {
	if !m.hdr {
		m.hdr = true
		if len(b) < 2 {
			return 0, io.ErrShortBuffer
		}
		b[0] = KindRread
		b[1] = m.Tag
		return 2, nil
	}
	if m.Data == nil {
		return 0, io.EOF
	}
	return m.Data.Read(b)
}

func (m *Rread) Write(b []byte) (int, error) {
	if len(b) < 1 {
		return 0, ErrBufTooSmall
	}
	m.Tag = b[0]
	m.Data = bytes.NewReader(b[1:])
	return len(b), nil
}

func (m *Rread) String() string {
	return fmt.Sprintf("Rread tag(%x)", m.Tag)
}
