// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

// Package msg is all the wire types and their packing.
package msg

import "fmt"

// ErrBufTooSmall is returned when a buffer is too small for a message to write
// itself into.
var ErrBufTooSmall = fmt.Errorf("msg: destination buffer too small")

// These are a bunch of constants for message types.
const (
	_ uint8 = iota
	KindRerror
	KindTping
	KindRping
	KindThello
	KindRhello
	KindTgoodbye
	KindRgoodbye /* not used */
	KindTauth0   /* auth messages not implemented */
	KindRauth0   /* auth messages not implemented */
	KindTauth1   /* auth messages not implemented */
	KindRauth1   /* auth messages not implemented */
	KindTread
	KindRread
	KindTwrite
	KindRwrite
	KindTsync
	KindRsync
	KindTmax

	maxStringSize = 1024
)

func fmtScore(s []byte) string {
	switch len(s) {
	case 0:
		return "nil"
	case 20:
		return fmt.Sprintf("sha1!%20x", s)
	}
	return fmt.Sprintf("???!%x", s)
}
