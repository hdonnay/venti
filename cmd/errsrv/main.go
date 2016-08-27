// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

// Errsrv returns an error for all messages besides the initial handshake and
// pings.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/hdonnay/venti"
	"github.com/hdonnay/venti/ventitest"
)

var (
	addr = flag.String("a", ":17034", "listen address")
)

func main() {
	flag.Parse()
	fs := ventitest.NewErrFS(fmt.Errorf("errsrv: this is what you wanted!"))
	log.Fatal(venti.ListenAndServe(*addr, fs.Handshake))
}
