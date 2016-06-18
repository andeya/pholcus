// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// taken from http://golang.org/src/pkg/net/ipraw_test.go

package ping

import (
	"testing"
)

func TestPing(t *testing.T) {

	t.Log(Ping("www.baidu.com", 5e9))
}
