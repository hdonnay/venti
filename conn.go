// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package venti

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/hdonnay/venti/internal/msg"
	"github.com/hdonnay/venti/internal/pack"
)

var (
	// ErrBadVersion is sent to the client if they attempt to initiate a
	// connection with no common version.
	ErrBadVersion = fmt.Errorf("no common version")

	// errUnexpectedHello is sent to the client if the sends an extra hello.
	errUnexpectedHello = fmt.Errorf("got Thello inside established connection")
	// errGoodbye is a non-error used to signal that a connection should be torn
	// down.
	errGoodbye = fmt.Errorf("goodbye")

	serverV = []byte("venti-04-hdonnay/venti\n")

	bufferPool sync.Pool
)

const (
	initBufSz = 4096
	versions  = "04"
)

func newBuffer() *bytes.Buffer {
	if v := bufferPool.Get(); v != nil {
		return v.(*bytes.Buffer)
	}
	return bytes.NewBuffer(make([]byte, 0, initBufSz))
}
func doneBuffer(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}

type conn struct {
	net.Conn
	// These have their own mutexes, just don't end-around them.
	r *pack.Dechunker
	w *pack.Chunker

	// This is the user-supplied function and the Handler derived from it.
	hs Handshake
	h  Handler
}

func accept(nc net.Conn, h Handshake) {
	c := &conn{
		Conn: nc,
		r:    pack.Dechunk(nc),
		w:    pack.Chunk(nc),
		hs:   h,
	}

	// The venti protocol starts with the exchanging of the strings.
	clientV, err := c.r.Line()
	if err != nil {
		c.Err(err)
		return
	}
	if _, err := c.Conn.Write(serverV); err != nil {
		c.Err(err)
		return
	}

	// The handshake function handles the initial Thello/Rhello messages.
	c.h, err = c.handshake(clientV)
	if err != nil {
		c.Err(err)
		return
	}

	for {
		if err := c.handle(); err != nil {
			c.Err(err)
			return
		}
	}
}

func (c *conn) handshake(cv string) (Handler, error) {
	ok := false
	for _, v := range ParseVersion(cv) {
		if strings.Contains(versions, v) {
			ok = true
			break
		}
	}
	if !ok {
		return nil, ErrBadVersion
	}

	th := &msg.Thello{}
	buf, err := c.readPacket()
	defer doneBuffer(buf)
	if err != nil {
		return nil, err
	}

	if k := buf.Next(1)[0]; k != msg.KindThello {
		return nil, fmt.Errorf("expected Thello, got %x", k)
	}

	if _, err := buf.ReadFrom(th); err != nil {
		return nil, err
	}

	r, h, err := c.hs(&Thello{
		UID:      th.UID,
		Strength: th.Strong,
		Crypto:   th.Crypto,
		Codec:    th.Codec,
	})
	if err != nil {
		return nil, err
	}

	if r == nil {
		r = &Rhello{
			SID: th.UID,
		}
	}

	rh := &msg.Rhello{
		Tag:    th.Tag,
		SID:    r.SID,
		Crypto: r.Crypto,
		Codec:  r.Codec,
	}

	out := c.w.New()
	defer out.Close()
	if _, err := io.Copy(out, rh); err != nil {
		return nil, err
	}
	return h, nil
}

// ReadPacket pulls the next packet off the connection.
//
// The returned *bytes.Buffer should be passed to doneBuffer() when done.
// A valid Buffer is always returned, and always needs to be done'd.
func (c *conn) readPacket() (*bytes.Buffer, error) {
	buf := newBuffer()
	if _, err := io.Copy(buf, c.r); err != nil {
		return buf, err
	}
	return buf, nil
}

func (c *conn) Close() error {
	// if we need to do more advance closing behavior, do it here
	return c.Conn.Close()
}

// Construct an Rerror packet and send it.
//
// If passed 'goodbye', end the connection.
func (c *conn) Err(e error) {
	if e == errGoodbye {
		c.Close()
		return
	}
	out := c.w.New()
	defer out.Close()
	io.Copy(out, &msg.Rerror{Err: e.Error()})
}

func (c *conn) handle() error {
	buf, err := c.readPacket()
	defer doneBuffer(buf)

	if err != nil {
		return err
	}

	var r io.Reader

	kind, err := buf.ReadByte()
	if err != nil {
		return err
	}
	switch kind {
	case msg.KindThello:
		return errUnexpectedHello
	case msg.KindTwrite:
		t := &msg.Twrite{}
		n, err := t.Write(buf.Bytes())
		if err != nil {
			return err
		}
		buf.Next(n)
		score, err := c.h.Write(Type(t.Type), buf)
		if err != nil {
			return err
		}
		r = &msg.Rwrite{
			Tag:   t.Tag,
			Score: score,
		}
	case msg.KindTread:
		t := &msg.Tread{}
		if _, err := t.Write(buf.Bytes()); err != nil {
			return err
		}
		rd, err := c.h.Read(Score(t.Score), Type(t.Type), int64(t.Count))
		defer func() {
			if pr, ok := rd.(PoolReader); ok {
				pr.Done()
			}
		}()
		if err != nil {
			return err
		}
		r = &msg.Rread{
			Tag:  t.Tag,
			Data: rd,
		}
	case msg.KindTsync:
		if err := c.h.Sync(); err != nil {
			return err
		}
		r = &msg.Rsync{Tag: buf.Next(1)[0]}
	case msg.KindTgoodbye:
		return errGoodbye
	case msg.KindTping:
		r = &msg.Rping{Tag: buf.Next(1)[0]}
	default:
		return fmt.Errorf("unexpected type %x", kind)
	}

	out := c.w.New()
	// This leaks a buffer if the Copy fails
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	return out.Close()
}
