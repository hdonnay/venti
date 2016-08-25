// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package msg

import (
	"fmt"
	"io"
)

// Tgoodbye is VtTgoodbye.
type Tgoodbye struct {
	Tag byte
}

func (m *Tgoodbye) Read(b []byte) (int, error) {
	if cap(b) < 2 {
		return 0, io.ErrShortBuffer
	}
	b = b[:3]
	b[0] = KindTgoodbye
	b[1] = m.Tag
	return 2, io.EOF
}

func (m *Tgoodbye) Write(b []byte) (int, error) {
	if len(b) < 1 {
		return 0, ErrBufTooSmall
	}
	m.Tag = b[0]
	return 1, nil
}

func (m *Tgoodbye) String() string {
	return fmt.Sprintf("Tgoodbye tag(%x)", m.Tag)
}
