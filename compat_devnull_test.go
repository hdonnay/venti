// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

// +build compat

package venti_test

import (
	"bytes"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hdonnay/venti"
)

func TestDevnullWrites(t *testing.T) {
	p := filepath.Join(tmp, port())

	dn := exec.Command(exes["devnull"], "-a", "unix!"+p)
	t.Log(dn.Args)
	if err := dn.Start(); err != nil {
		t.Fatal(err)
	}
	defer dn.Process.Signal(os.Interrupt)
	t.Logf("devnull spawned as %d\n", dn.Process.Pid)
	time.Sleep(500 * time.Millisecond) // wait for the server to start

	conn, err := net.Dial("unix", p)
	if err != nil {
		t.Logf("devnull state: %v\n", dn.Process)
		t.Fatal(err)
	}

	c, err := venti.NewClient(conn)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	_, rd, h := randomBlock()
	s, err := c.Write(venti.VtData, rd)
	if err != nil {
		t.Fatal(err)
	}

	if exp, got := h.Sum(nil), s; !bytes.Equal(exp, got) {
		t.Fatalf("wrong score, exp: %v got: %v\n", exp, got)
	}
	t.Logf("exp: %v got: %v\n", venti.Score(h.Sum(nil)), s)
}

func TestDevnullReads(t *testing.T) {
	p := filepath.Join(tmp, port())

	dn := exec.Command(exes["devnull"], "-a", "unix!"+p)
	t.Log(dn.Args)
	if err := dn.Start(); err != nil {
		t.Fatal(err)
	}
	defer dn.Process.Signal(os.Interrupt)
	t.Logf("devnull spawned as %d\n", dn.Process.Pid)
	time.Sleep(500 * time.Millisecond) // wait for the server to start

	conn, err := net.Dial("unix", p)
	if err != nil {
		t.Logf("devnull state: %v\n", dn.Process)
		t.Fatal(err)
	}

	c, err := venti.NewClient(conn)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	_, err = c.Read(venti.VtData, venti.Score(make([]byte, 20)), 0)
	if err == nil {
		t.Fatal("expected an error, didn't get one")
	}
	t.Log(err)
}
