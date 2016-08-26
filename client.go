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

func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewClient(conn)
}

func NewClient(conn net.Conn) (*Client, error) {
	c := &Client{
		Conn: conn,
		w:    pack.Chunk(conn),
		r:    pack.Dechunk(conn),
		done: make(chan struct{}),

		ts: &tagset{},
	}
	//c.w.Chatty = true
	//c.r.Chatty = true
	var err error
	c.Version, err = c.version()
	if err != nil {
		conn.Close()
		return nil, err
	}
	go c.recv()
	if err := c.hello(); err != nil {
		conn.Close()
		return nil, err
	}

	return c, nil
}

type Client struct {
	net.Conn
	w    *pack.Chunker
	r    *pack.Dechunker
	done chan struct{}
	pool sync.Pool
	err  error

	ts *tagset

	Version string
}

func (c *Client) recv() {
	for {
		select {
		case <-c.done:
			return
		default:
		}
		buf := c.newBuf()
		if _, err := io.Copy(buf, c.r); err != nil {
			c.err = err
			return
		}
		tag := buf.Bytes()[1]
		// Send is a channel send and clunk
		c.ts.Send(tag, buf)
	}
}

func (c *Client) version() (string, error) {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "venti-%s-%s\n", strings.Join(vers, ":"), verComment)
	if _, err := io.Copy(c.Conn, buf); err != nil {
		return "", err
	}
	srvV, err := c.r.Line()
	if err != nil {
		return "", err
	}
	cm := ""
	vs := ParseVersion(srvV)
	// walk backwards so we get the biggest number, in the future with jetpacks
	// and multiple numbers.
Cmp:
	for i := len(vs) - 1; i >= 0; i-- {
		for _, v := range vers {
			if v == vs[i] {
				cm = v
				break Cmp
			}
		}
	}
	if cm == "" {
		defer c.Close()
		return "", fmt.Errorf("no common version")
	}
	return cm, nil
}

func (c *Client) hello() error {
	tag, r := c.ts.New()

	t := &msg.Thello{
		Tag:     tag,
		Version: c.Version,
		UID:     "anonymous",
	}
	w := c.w.New()
	if _, err := io.Copy(w, t); err != nil {
		return err
	}
	w.Close()

	buf := <-r
	defer c.doneBuf(buf)

	if err := want(msg.KindRhello, buf); err != nil {
		return err
	}
	h := &msg.Rhello{}
	if _, err := io.Copy(h, buf); err != nil {
		return err
	}
	// if negotiating stuff, do that here.

	return nil
}

func (c *Client) goodbye() error {
	tag, _ := c.ts.New()
	t := &msg.Tgoodbye{Tag: tag}
	w := c.w.New()
	defer w.Close()
	if _, err := io.Copy(w, t); err != nil {
		c.ts.Clunk(tag)
		return err
	}
	return nil
}

func (c *Client) Read(t Type, s Score) (io.Reader, error) {
	return nil, nil
}

func (c *Client) Write(t Type, r io.Reader) (Score, error) {
	return nil, nil
}

func (c *Client) Ping() error {
	tag, r := c.ts.New()

	t := &msg.Tping{Tag: tag}
	w := c.w.New()
	if _, err := io.Copy(w, t); err != nil {
		c.ts.Clunk(tag)
		return err
	}
	w.Close()

	buf := <-r
	defer c.doneBuf(buf)

	if err := want(msg.KindRping, buf); err != nil {
		return err
	}
	// Don't bother reading the packet, if we've gotten here, it's all good.
	return nil
}

func (c *Client) Sync() error {
	tag, r := c.ts.New()

	t := &msg.Tsync{Tag: tag}
	w := c.w.New()
	if _, err := io.Copy(w, t); err != nil {
		c.ts.Clunk(tag)
		return err
	}
	w.Close()

	buf := <-r
	defer c.doneBuf(buf)

	if err := want(msg.KindRsync, buf); err != nil {
		return err
	}
	// Don't bother reading the packet, if we've gotten here, it's all good.
	return nil
}

func (c *Client) Close() error {
	c.goodbye()
	close(c.done)
	return c.Conn.Close()
}

func (c *Client) newBuf() *bytes.Buffer {
	if b := c.pool.Get(); b != nil {
		return b.(*bytes.Buffer)
	}
	return bytes.NewBuffer(make([]byte, 0, initBufSz))
}
func (c *Client) doneBuf(b *bytes.Buffer) {
	b.Reset()
	c.pool.Put(b)
}

type tagset struct {
	sync.Mutex
	next uint8
	wait [256]chan *bytes.Buffer
}

func (t *tagset) New() (uint8, chan *bytes.Buffer) {
	// There might be a problem if we attempt more than 256 requests in-flight
	t.Lock()
	defer t.Unlock()
	lp := t.next
	for t.next++; t.next != lp; t.next++ {
		if t.wait[t.next] == nil {
			ch := make(chan *bytes.Buffer)
			t.wait[t.next] = ch
			return t.next, ch
		}
	}
	panic("too many concurrent requests") // time to rewrite this if you're here
}

func (t *tagset) Tag(tg uint8) chan *bytes.Buffer {
	t.Lock()
	defer t.Unlock()
	return t.wait[tg]
}

func (t *tagset) Clunk(tg uint8) {
	t.Lock()
	defer t.Unlock()
	close(t.wait[tg])
	t.wait[tg] = nil
}

func (t *tagset) Send(tg uint8, buf *bytes.Buffer) {
	t.Lock()
	defer t.Unlock()
	t.wait[tg] <- buf
	close(t.wait[tg])
	t.wait[tg] = nil
}

func want(w byte, buf *bytes.Buffer) error {
	h := buf.Next(1)[0]
	switch h {
	case msg.KindRerror:
		r := &msg.Rerror{}
		io.Copy(r, buf)
		return r
	case w:
		break
	default:
		return fmt.Errorf("incoming message: wanted %x, got %x", w, h)
	}
	return nil
}
