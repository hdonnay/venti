// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package msg

import (
	"fmt"
	"io"

	"github.com/hdonnay/venti/internal/pack"
)

// Rerror is VtRerror.
type Rerror struct {
	Tag byte
	Err string
}

func (m *Rerror) Read(b []byte) (int, error) {
	sz := len(m.Err) + 4
	if cap(b) < sz {
		return 0, io.ErrShortBuffer
	}
	b = b[:sz]
	b[0] = KindRerror
	b[1] = m.Tag
	pack.String(b[2:], m.Err)
	return sz, io.EOF
}

func (m *Rerror) Write(b []byte) (int, error) {
	if len(b) < 3 {
		return 0, ErrBufTooSmall
	}
	m.Tag = b[0]
	_, m.Err = pack.UnString(b[1:])
	return 3 + len(m.Err), nil
}

func (m *Rerror) String() string {
	return fmt.Sprintf("Rerror tag(%x) %q", m.Tag, m.Err)
}

func (m *Rerror) Error() string {
	return m.Err
}
