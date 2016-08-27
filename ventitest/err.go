// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package ventitest

import (
	"io"

	"github.com/hdonnay/venti"
)

func NewErrFS(err error) *ErrFS {
	return &ErrFS{Err: err}
}

type ErrFS struct {
	Err error
}

func (e *ErrFS) Handshake(_ *venti.Thello) (*venti.Rhello, venti.Handler, error) {
	return nil, e, nil
}

func (e *ErrFS) Write(_ venti.Type, _ io.Reader) (venti.Score, error) {
	return venti.Score(nil), e.Err
}
func (e *ErrFS) Read(_ venti.Score, _ venti.Type, _ int64) (io.Reader, error) {
	return nil, e.Err
}

func (e *ErrFS) Sync() error {
	return e.Err
}
