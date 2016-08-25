// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package msg

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"
)

var Pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

type RRow struct {
	R interface {
		io.Reader
		fmt.Stringer
	}
	B []byte
}

func readerTest(t *testing.T, tbl []RRow) {
	by := Pool.Get().([]byte)
	defer Pool.Put(by)
	buf := bytes.NewBuffer(by)

	for _, row := range tbl {
		buf.Reset()
		t.Logf("packing: %s\n", row.R)

		_, err := io.Copy(buf, row.R)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(buf.Bytes(), row.B) {
			t.Logf("have: %v\n", buf.Bytes())
			t.Logf("want: %v\n", row.B)
			t.Fail()
		}
	}
}

type WRow struct {
	A, B io.ReadWriter
}

func writerTest(t *testing.T, tbl []WRow) {
	by := Pool.Get().([]byte)
	defer Pool.Put(by)
	buf := bytes.NewBuffer(by)

	for _, row := range tbl {
		buf.Reset()

		_, err := io.Copy(buf, row.A)
		if err != nil {
			t.Fatal(err)
		}
		buf.Next(1) // chop off type header

		_, err = io.Copy(row.B, buf)
		if err != nil {
			t.Fatalf("%T %v", row.B, err)
		}

		a, b := reflect.TypeOf(row.A).Elem(), reflect.TypeOf(row.B).Elem()
		for i := 0; i < a.NumField(); i++ {
			av, bv := a.Field(i), b.Field(i)
			if !reflect.DeepEqual(av, bv) {
				t.Fatalf("found unequal: %v != %v", av, bv)
			}
		}
	}
}
