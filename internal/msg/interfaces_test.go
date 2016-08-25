// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package msg_test

import (
	"fmt"
	"io"

	"github.com/hdonnay/venti/internal/msg"
)

var (
	_ io.ReadWriter = (*msg.Thello)(nil)
	_ fmt.Stringer  = (*msg.Thello)(nil)
	_ io.ReadWriter = (*msg.Rhello)(nil)
	_ fmt.Stringer  = (*msg.Rhello)(nil)

	_ io.ReadWriter = (*msg.Tping)(nil)
	_ fmt.Stringer  = (*msg.Tping)(nil)
	_ io.ReadWriter = (*msg.Rping)(nil)
	_ fmt.Stringer  = (*msg.Rping)(nil)

	_ io.ReadWriter = (*msg.Tread)(nil)
	_ fmt.Stringer  = (*msg.Tread)(nil)
	_ io.ReadWriter = (*msg.Rread)(nil)
	_ fmt.Stringer  = (*msg.Rread)(nil)

	_ io.ReadWriter = (*msg.Twrite)(nil)
	_ fmt.Stringer  = (*msg.Twrite)(nil)
	_ io.ReadWriter = (*msg.Rwrite)(nil)
	_ fmt.Stringer  = (*msg.Rwrite)(nil)

	_ io.ReadWriter = (*msg.Tsync)(nil)
	_ fmt.Stringer  = (*msg.Tsync)(nil)
	_ io.ReadWriter = (*msg.Rsync)(nil)
	_ fmt.Stringer  = (*msg.Rsync)(nil)

	_ io.ReadWriter = (*msg.Tgoodbye)(nil)
	_ fmt.Stringer  = (*msg.Tgoodbye)(nil)

	_ io.ReadWriter = (*msg.Rerror)(nil)
	_ fmt.Stringer  = (*msg.Rerror)(nil)
	_ error         = (*msg.Rerror)(nil)
)
