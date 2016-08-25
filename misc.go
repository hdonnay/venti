// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package venti

import "strings"

// ParseVersion returns the versions out of a venti version string.
func ParseVersion(ver string) []string {
	vs := strings.SplitN(ver, "-", 3)[1]
	return strings.Split(vs, ":")
}
