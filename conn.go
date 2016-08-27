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

	bufferPool sync.Pool
)

const (
	initBufSz = 4096
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
	defer c.Close()

	// The venti protocol starts with the exchanging of the strings.
	clientV, err := c.r.Line()
	if err != nil {
		return
	}
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "venti-%s-%s\n", strings.Join(vers, ":"), verComment)
	if _, err := io.Copy(c.Conn, buf); err != nil {
		return
	}

	// The handshake function handles the initial Thello/Rhello messages.
	c.h, err = c.handshake(clientV)
	if err != nil {
		return
	}

	for {
		if err := c.handle(); err != nil {
			return
		}
	}
}

func (c *conn) handshake(cv string) (Handler, error) {
	ok := false
	for _, v := range ParseVersion(cv) {
		if strings.Contains(strings.Join(vers, ""), v) {
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
		err := fmt.Errorf("expected Thello, got %x", k)
		c.Err(buf.Next(1)[0], err)
		return nil, err
	}

	if _, err := io.Copy(th, buf); err != nil {
		c.Err(th.Tag, err)
		return nil, err
	}

	r, h, err := c.hs(&Thello{
		UID:      th.UID,
		Strength: th.Strong,
		Crypto:   th.Crypto,
		Codec:    th.Codec,
	})
	if err != nil {
		c.Err(th.Tag, err)
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
		c.Err(th.Tag, err)
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
func (c *conn) Err(tag uint8, e error) {
	if e == errGoodbye {
		c.Close()
		return
	}
	out := c.w.New()
	defer out.Close()
	io.Copy(out, &msg.Rerror{
		Tag: tag,
		Err: e.Error(),
	})
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
		c.Err(buf.Next(1)[0], errUnexpectedHello)
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
			c.Err(t.Tag, err)
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
		if rc, ok := rd.(io.ReadCloser); ok {
			defer rc.Close()
		}
		if err != nil {
			c.Err(t.Tag, err)
			return err
		}
		r = &msg.Rread{
			Tag:  t.Tag,
			Data: rd,
		}
	case msg.KindTsync:
		tag := buf.Next(1)[0]
		if err := c.h.Sync(); err != nil {
			c.Err(tag, err)
			return err
		}
		r = &msg.Rsync{Tag: tag}
	case msg.KindTgoodbye:
		return errGoodbye
	case msg.KindTping:
		r = &msg.Rping{Tag: buf.Next(1)[0]}
	default:
		err := fmt.Errorf("unexpected type %x", kind)
		c.Err(buf.Next(1)[0], err)
		return err
	}

	out := c.w.New()
	// This leaks a buffer if the Copy fails
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	return out.Close()
}
