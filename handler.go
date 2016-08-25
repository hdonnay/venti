// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package venti

import "io"

// Thello is the client's hello message.
//
// All current implementations ignore all these fields.
type Thello struct {
	// UID is the identity of the connecting user.
	UID string
	// Strength is flags to negotiate authentication, encryption, and
	// compression.
	Strength uint8

	// Crypto and Codec are arguments to the negotiation indicated by Strength
	Crypto []byte
	Codec  []byte
}

// Rhello is the server's response to a Thello.
//
// All current implementations ignore all these fields.
type Rhello struct {
	// SID is the identity of the server.
	SID string

	// Crypto and Codec are the response to the crypto and compression
	// negotiation.
	Crypto uint8
	Codec  uint8
}

// Handshake takes the parameters sent by the client's hello, and returns
// a Handler if it can continue, or an error otherwise.
//
// If a nil is returned instead of an Rhello struct, a default handshaking
// function will be used.
type Handshake func(*Thello) (*Rhello, Handler, error)

// Handler is the interface a venti server implementation must provide.
//
// Semantics documented here take precidence over those described in venti(7).
type Handler interface {
	// Read requests the block identified by the provided score, type pair. The
	// count argument is the maximum expected size of the response.
	//
	// If the supplied io.Reader satifies the PoolReader interface, Put will be
	// called when its consumed.
	Read(score Score, kind Type, count int64) (io.Reader, error)

	// Write requests that the provided data of the given type be written.
	// The returned score should be a hash of the data using the negotiated
	// algorithm.
	//
	// The passed-in io.Reader is only valid for the length of the function
	// call and should not be retained.
	Write(kind Type, data io.Reader) (Score, error)

	// Sync is called when the client has requested that the previous writes be
	// persisted. It should delay returning until this is done.
	Sync() error
}

// PoolReader is used to recycle the io.Reader returned by (Handler).Read if
// possible.
type PoolReader interface {
	io.Reader
	Done()
}
