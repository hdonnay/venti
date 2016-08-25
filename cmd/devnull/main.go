// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

/*
Devnull is a reimplementation of src/venti/cmd/devnull.c.

It accepts and discards all writes, and returns errors on all reads.
*/
package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"hash"
	"io"
	"net"
	"os"

	"github.com/hdonnay/venti"
)

var (
	addr = flag.String("a", ":17034", "listen address")
	v    = flag.Bool("V", false, "verbose")
)

func main() {
	flag.Parse()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := venti.Serve(l, NewFS); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// NewFS is the venti.Handshake for our devnull server.
func NewFS(_ *venti.Thello) (*venti.Rhello, venti.Handler, error) {
	return nil, &fs{score: sha1.New()}, nil
}

type fs struct {
	score hash.Hash
}

// Read returns errors.
func (f *fs) Read(_ venti.Score, _ venti.Type, _ int64) (io.Reader, error) {
	return nil, fmt.Errorf("no such block")
}

// Write calculates and returns the score of the block, and then does nothing
// with it.
func (f *fs) Write(_ venti.Type, r io.Reader) (venti.Score, error) {
	defer f.score.Reset()
	if _, err := io.Copy(f.score, r); err != nil {
		return nil, err
	}
	s := venti.Score(f.score.Sum(nil))
	if *v {
		fmt.Fprintf(os.Stderr, "discarded block with score %v\n", s)
	}
	return s, nil
}

// Sync is a noop.
func (f *fs) Sync() error { return nil }
