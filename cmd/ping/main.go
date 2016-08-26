// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

/*
Ping pings a venti server.
*/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hdonnay/venti"
)

var (
	addr = flag.String("a", "[::1]:17034", "listen address")
)

func main() {
	flag.Parse()

	c, err := venti.Dial(*addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer c.Close()

	if err := c.Ping(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
