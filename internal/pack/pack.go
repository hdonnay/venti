// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

// Package pack is the message packing scheme for venti, and some convenience
// functions for reading/writing venti packets.
//
// See venti(7) from Plan 9 or plan9port.
package pack

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"sync"
)

var be = binary.BigEndian

// UnString unpacks a string from the []byte, returning the number of bytes used
// and the string.
func UnString(b []byte) (int, string) {
	sz := int(be.Uint16(b[:2]))
	return sz + 2, string(b[2 : 2+sz])
}

// String writes the string into the []byte, returning the number of bytes used.
func String(b []byte, s string) int {
	if len(s) > 1024 {
		panic("string larger than venti protocol supports")
	}
	be.PutUint16(b[:2], uint16(len(s)))
	copy(b[2:], s)
	return len(s) + 2
}

// UnUint16 unpacks a string from the []byte, returning the number of bytes used
// and the uint16.
func UnUint16(b []byte) (int, uint16) {
	return 2, be.Uint16(b[:2])
}

// Uint16 writes the uint16 into the []byte, returning the number of bytes used.
func Uint16(b []byte, v uint16) int {
	be.PutUint16(b[:2], v)
	return 2
}

// UnUint32 unpacks a string from the []byte, returning the number of bytes used
// and the uint32.
func UnUint32(b []byte) (int, uint32) {
	return 4, be.Uint32(b[:4])
}

// Uint32 writes the uint32 into the []byte, returning the number of bytes used.
func Uint32(b []byte, v uint32) int {
	be.PutUint32(b[:4], v)
	return 4
}

// UnUint64 unpacks a string from the []byte, returning the number of bytes used
// and the uint64.
func UnUint64(b []byte) (int, uint64) {
	return 8, be.Uint64(b[:8])
}

// Uint64 writes the uint64 into the []byte, returning the number of bytes used.
func Uint64(b []byte, v uint64) int {
	be.PutUint64(b[:8], v)
	return 8
}

// Var copies the []byte v into []byte b, returning the number of bytes used.
func Var(b []byte, v []byte) int {
	if len(v) > 255 { // max uint8
		panic("[]byte larger and venti protocol supports")
	}
	b[0] = uint8(len(v))
	copy(b[1:], v)
	return len(v) + 1
}

// UnVar unpacks a []byte from the []byte b. The returned slice shares the
// backing array, and the int is the number of bytes used.
func UnVar(b []byte) (int, []byte) {
	l := int(uint8(b[0]))
	return l + 1, b[1 : l+1]
}

var (
	pktPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 4096))
		},
	}
)

// Chunker yeilds writers that emit venti-format packets.
//
// We only emit venti v04 compatable packets.
type Chunker struct {
	sync.Mutex
	w io.Writer
}

// cw buffers all writes until closed, then determines the length written,
// secures the lock, and writes the buffered data.
type cw struct {
	bytes.Buffer
	Mu *sync.Mutex
	W  io.Writer
}

func (w *cw) Close() error {
	defer pktPool.Put(&w.Buffer)
	b := make([]byte, 4)

	be.PutUint32(b, uint32(w.Buffer.Len()))
	_, err := io.Copy(w.W, io.MultiReader(bytes.NewReader(b), &w.Buffer))
	return err
}

// Chunk returns a Chunker that writes to 'w'.
func Chunk(w io.Writer) *Chunker {
	return &Chunker{
		w: w,
	}
}

// New returns a io.WriteCloser that buffers writes and frames and flushes data
// when closed.
func (w *Chunker) New() io.WriteCloser {
	b := pktPool.Get().(*bytes.Buffer)
	return &cw{
		Buffer: *b,
		Mu:     &w.Mutex,
		W:      w.w,
	}
}

// Dechunker reads venti-format packets.
// It strips the leading length and inserts io.EOFs after each.
//
// We only support venti v04 compatable packets.
type Dechunker struct {
	sync.Mutex
	r      *bufio.Reader
	remain int
}

// Dechunk return a Dechunker reading from 'r'.
func Dechunk(r io.Reader) *Dechunker {
	return &Dechunker{
		r:      bufio.NewReader(r),
		remain: -1,
	}
}

// Line reads until the next '\n'. Used for the initial version exchange.
func (d *Dechunker) Line() (string, error) {
	d.Lock()
	defer d.Unlock()
	return d.r.ReadString('\n')
}

// Read reads until the end of the message.
//
// A Read call after an io.EOF is returned will read the next packet.
func (d *Dechunker) Read(b []byte) (int, error) {
	d.Lock()
	defer d.Unlock()
	if d.remain == 0 {
		d.remain--
		return 0, io.EOF
	}
	if d.remain < 0 {
		// Use the buffer passed in as scratch space to determine the length of
		// the message.
		sz := b[:4]
	Try:
		n, err := d.r.Read(sz)
		if err != nil {
			return n, err
		}
		if n < 4 {
			sz = sz[n:]
			goto Try
		}
		d.remain = int(binary.BigEndian.Uint32(b[:4]))
	}

	if len(b) > d.remain {
		b = b[:d.remain]
	}

	n, err := d.r.Read(b)
	d.remain -= n
	return n, err
}
