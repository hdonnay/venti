#!/bin/sh

cd "$(git rev-parse --show-toplevel)"

find -type f -name '*.go' | while read f; do
	if sed '1q' "$f" | grep -vq '^// Copyright' ; then
		echo $f
		cat <<EOF > "${f}_"
// Copyright 2016 The Venti Authors. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

EOF
		cat "$f" >> "${f}_"
		mv "${f}_" "${f}"
	fi
done
