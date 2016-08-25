// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package msg

import (
	"fmt"
	"io"

	"github.com/hdonnay/venti/internal/pack"
)

// Thello initiates the handshake.
//
// Cribbing from the venti(7) manpage:
//
//	Venti connections must begin with a hello transaction.  The
//	VtThello message contains the protocol version that the
//	client has chosen to use.  The fields strength, crypto, and
//	codec could be used to add authentication, encryption, and
//	compression to the Venti session but are currently ignored.
//	The rcrypto, and rcodec fields in the VtRhello response are
//	similarly ignored.  The uid and sid fields are intended to
//	be the identity of the client and server but, given the lack
//	of authentication, should be treated only as advisory.  The
//	initial hello should be the only hello transaction during
//	the session.
type Thello struct {
	Tag     byte
	Version string
	UID     string
	Strong  byte
	Crypto  []byte
	Codec   []byte
}

func (m *Thello) Read(b []byte) (int, error) {
	sz := 9 + len(m.Version) + len(m.UID) + len(m.Codec) + len(m.Crypto)
	if cap(b) < sz {
		return 0, io.ErrShortBuffer
	}
	b = b[:sz]

	b[0] = KindThello
	b[1] = m.Tag
	off := 2
	off += pack.String(b[off:], m.Version)
	off += pack.String(b[off:], m.UID)
	b[off] = m.Strong
	off++
	off += pack.Var(b[off:], m.Crypto)
	off += pack.Var(b[off:], m.Codec)

	return off, io.EOF
}

func (m *Thello) String() string {
	return fmt.Sprintf("Thello tag(%x) v%q uid(%s) strong?%v crypto(%x) codec(%x)",
		m.Tag, m.Version, m.UID, m.Strong != 0, m.Crypto, m.Codec)
}

func (m *Thello) Write(b []byte) (int, error) {
	var off, acc int = 1, 0
	if len(b) < 8 {
		return 0, ErrBufTooSmall
	}
	m.Tag = b[0]
	// accoring to venti(7), most of this is ignored.
	acc, m.Version = pack.UnString(b[1:])
	off += acc
	acc, m.UID = pack.UnString(b[off:])
	off += acc
	m.Strong = b[off]
	off++
	acc, m.Crypto = pack.UnVar(b[off:])
	off += acc
	acc, m.Codec = pack.UnVar(b[off:])
	return off + acc, nil
}

// Rhello is the server response to Thello. It completes the venti handshake.
type Rhello struct {
	Tag    byte
	SID    string
	Crypto byte
	Codec  byte
}

func (m *Rhello) String() string {
	return fmt.Sprintf("Rhello tag(%x) sid(%s) crypto(%x) codec(%x)",
		m.Tag, m.SID, m.Crypto, m.Codec)
}

func (m *Rhello) Read(b []byte) (int, error) {
	sz := 6 + len(m.SID)
	if cap(b) < sz {
		return 0, io.ErrShortBuffer
	}
	b = b[:sz]

	b[0] = KindRhello
	b[1] = m.Tag
	off := 2

	off += pack.String(b[off:], m.SID)

	b[off] = m.Crypto
	off++
	b[off] = m.Codec

	return sz, io.EOF
}

func (m *Rhello) Write(b []byte) (int, error) {
	var off, acc int = 1, 0
	if len(b) < 3 {
		return 0, ErrBufTooSmall
	}
	m.Tag = b[0]
	acc, m.SID = pack.UnString(b[1:])
	off += acc
	m.Crypto = b[off]
	m.Codec = b[off+1]
	return off + 2, nil
}
