// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package ventitest

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"sync"

	"github.com/hdonnay/venti"
)

var zeroScore = make([]byte, 20)

func NewMemFS() *MemFS {
	return &MemFS{
		mu:    &sync.RWMutex{},
		block: make(map[string][]byte),
		h:     sha1.New(),
	}
}

type MemFS struct {
	mu    *sync.RWMutex
	block map[string][]byte
	h     hash.Hash
}

func (fs *MemFS) Reset() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.block = make(map[string][]byte)
}

func (fs *MemFS) Handshake(t *venti.Thello) (*venti.Rhello, venti.Handler, error) {
	return nil, fs, nil
}

func (fs *MemFS) Write(_ venti.Type, r io.Reader) (venti.Score, error) {
	buf := &bytes.Buffer{}
	fs.mu.Lock()
	defer fs.mu.Unlock()
	defer fs.h.Reset()
	if _, err := io.Copy(buf, io.TeeReader(r, fs.h)); err != nil {
		return zeroScore, err
	}

	s := venti.Score(fs.h.Sum(nil))
	fs.block[s.String()] = buf.Bytes()
	return s, nil
}

func (fs *MemFS) Read(s venti.Score, _ venti.Type, ct int64) (io.Reader, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	b, ok := fs.block[s.String()]
	if !ok {
		return nil, fmt.Errorf("no such block")
	}
	return bytes.NewReader(b), nil
}

func (fs *MemFS) Sync() error {
	fs.mu.Lock()
	fs.mu.Unlock()
	return nil
}
