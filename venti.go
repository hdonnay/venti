// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

// Package venti is a group of libraries for writing venti(7) servers.
package venti

import "fmt"

var vers = []string{"04"}

const verComment = "hdonnay/venti"

// MaxFileSize is the biggest file venti supports.
const MaxFileSize = (1 << 48) - 1

// Score is the hash of a block.
//
// Currently the score function clients recognize is SHA1.
type Score []byte

func (s Score) String() string {
	switch len(s) {
	case 20:
		return fmt.Sprintf("sha1!%20x", string(s))
	case 0:
		return "nil"
	}
	return fmt.Sprintf("???!%x", string(s))
}

// Entry is a stub from libventi.h
type Entry struct {
	Gen        uint32
	PoiterSize uint32
	DataSize   uint32
	Type       uint8
	Flags      uint8
	Size       uint64
	Score      Score
}

// Root is a stub from libventi.h
type Root struct {
	Name      string
	Type      []byte
	Score     Score
	BlockSize uint32
	Prev      Score
}

// Type is a stub from libventi.h
type Type byte

// These are the Type constants.
const (
	VtData Type = iota << 3
	VtDir
	VtRoot
)

//go:generate stringer -type Type
