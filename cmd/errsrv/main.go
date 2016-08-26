// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

// Errsrv returns an error for all messages besides the initial handshake and
// pings.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/hdonnay/venti"
)

var (
	addr = flag.String("a", ":17034", "listen address")

	esErr = fmt.Errorf("errsrv: this is what you wanted!")
)

func main() {
	flag.Parse()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(venti.Serve(l, fs))
}

func fs(_ *venti.Thello) (*venti.Rhello, venti.Handler, error) {
	return nil, errs{}, nil
}

type errs struct{}

func (errs) Write(_ venti.Type, _ io.Reader) (venti.Score, error) {
	return venti.Score(nil), esErr
}
func (errs) Read(_ venti.Score, _ venti.Type, _ int64) (io.Reader, error) {
	return nil, esErr
}

func (errs) Sync() error {
	return esErr
}
