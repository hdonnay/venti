// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package venti_test

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"net"
	"testing"

	"github.com/hdonnay/venti"
	"github.com/hdonnay/venti/ventitest"
)

// This file has bits common to the normal tests and the compatability tests.

const (
	errStr = "could not"
)

var (
	errs = ventitest.NewErrFS(fmt.Errorf(errStr))
)

func randomBlock() (int64, io.Reader, hash.Hash) {
	sz := rand.Int63n(1 << 15)
	h := sha1.New()
	r := rand.New(rand.NewSource(rand.Int63()))
	return sz, io.TeeReader(io.LimitReader(r, sz), h), h
}

// Start a server on a random port, connect to it with a client, and return the
// client and a cleanup function.
func startServer(t *testing.T, h venti.Handshake) (*venti.Client, func()) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	go venti.Serve(l, h)
	t.Log("serving on", l.Addr())
	c, err := venti.Dial(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	return c, func() {
		l.Close()
		c.Close()
	}
}
