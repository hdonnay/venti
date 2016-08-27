// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package venti_test

import (
	"bytes"
	"crypto/sha1"
	"io"
	"testing"

	"github.com/hdonnay/venti"
	"github.com/hdonnay/venti/ventitest"
)

func TestWrite(t *testing.T) {
	memfs := ventitest.NewMemFS()

	c, done := startServer(t, memfs.Handshake)
	defer done()

	if err := c.Ping(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 50; i++ {
		_, r, h := randomBlock()
		sc, err := c.Write(venti.VtData, r)
		if err != nil {
			t.Fatal(err)
		}
		if exp, got := h.Sum(nil), sc; !bytes.Equal(exp, got) {
			t.Fatalf("exp: %020x, got: %020x", exp, got)
		}
		t.Logf("wrote %v\n", sc)
	}
	if err := c.Sync(); err != nil {
		t.Fatal(err)
	}
}

func TestReadWrite(t *testing.T) {
	memfs := ventitest.NewMemFS()
	var scs []venti.Score

	c, done := startServer(t, memfs.Handshake)
	defer done()

	if err := c.Ping(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 50; i++ {
		_, r, h := randomBlock()
		sc, err := c.Write(venti.VtData, r)
		if err != nil {
			t.Fatal(err)
		}
		if exp, got := h.Sum(nil), sc; !bytes.Equal(exp, got) {
			t.Fatalf("exp: %020x, got: %020x", exp, got)
		}
		t.Logf("wrote %v\n", sc)
		scs = append(scs, sc)
	}
	if err := c.Sync(); err != nil {
		t.Fatal(err)
	}

	h := sha1.New()
	for _, sc := range scs {
		r, err := c.Read(venti.VtData, sc, 4096)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(h, r); err != nil {
			t.Fatal(err)
		}
		if exp, got := sc, h.Sum(nil); !bytes.Equal(exp, got) {
			t.Fatalf("exp: %020x, got: %020x", exp, got)
		}
		t.Logf("read  %v\n", sc)

		r.Close()
		h.Reset()
	}
}

func TestWriteFail(t *testing.T) {
	c, done := startServer(t, errs.Handshake)
	defer done()

	_, err := c.Write(venti.VtData, &bytes.Buffer{})
	if err == nil {
		t.Fatal("wanted an error, didn't get one")
	}
	t.Log(err)
}

func TestReadFail(t *testing.T) {
	c, done := startServer(t, errs.Handshake)
	defer done()

	_, err := c.Read(venti.VtData, venti.Score(nil), 0)
	if err == nil {
		t.Fatal("wanted an error, didn't get one")
	}
	t.Log(err)
}

func TestSyncFail(t *testing.T) {
	c, done := startServer(t, errs.Handshake)
	defer done()

	err := c.Sync()
	if err == nil {
		t.Fatal("wanted an error, didn't get one")
	}
	t.Log(err)
}
