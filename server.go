// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package venti

import (
	"fmt"
	"net"
)

// Serve accepts connections on l and runs the supplied Handshake function and,
// if successful, the returned Handler.
func Serve(l net.Listener, h Handshake) error {
	if h == nil {
		return fmt.Errorf("venti: bad handshake function")
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go accept(conn, h)
	}
}
