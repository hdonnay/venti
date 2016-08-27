// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

// +build compat

/*
This is a test suite that's run when a "compat" tag is present.

It expects that you have a plan9port install at $PLAN9.
*/
package venti_test

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync/atomic"
	"testing"
)

var (
	exes = map[string]string{
		"venti":   "$PLAN9/bin/venti/venti",
		"devnull": "$PLAN9/src/cmd/venti/o.devnull",
		"read":    "$PLAN9/bin/venti/read",
	}
	nport = new(uint32)

	tmp string
)

func TestMain(m *testing.M) {
	prefix := "test setup: "
	flag.Parse()
	*nport = 17034
	var err error

	if os.Getenv("PLAN9") == "" {
		log.Fatal(prefix, "$PLAN9 unset")
	}

	tmp, err = ioutil.TempDir("", "venti-compat-test-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	mkdevnull()

	for k, v := range exes {
		exes[k] = os.ExpandEnv(v)
		if _, err := exec.LookPath(exes[k]); err != nil {
			log.Fatalf("%sunable to find %q at %q\n", prefix, k, exes[k])
		}
	}

	os.Exit(m.Run())
}

// the devnull venti server isn't built by default, go build it.
func mkdevnull() {
	dir := os.ExpandEnv("$PLAN9/src/cmd/venti")
	for _, a := range [][]string{
		{"9c", "devnull.c"},
		{"9l", "-o", "o.devnull", "devnull.o"},
	} {
		cmd := exec.Command(a[0], a[1:]...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}
}

// Grab a port we can listen to.
func port() string {
	return strconv.Itoa(int(atomic.AddUint32(nport, 1)))
}
