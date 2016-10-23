//
// Copyright yutopp 2016 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import (
	"testing"
)

func TestNormalizePath(t *testing.T) {
	{
		a := NormalizePath("/tmp", "/etc/hoge")
		if a != "/etc/hoge" {
			t.Fatalf("%s should be /etc/hoge", a)
		}
	}

	{
		a := NormalizePath("/tmp", "./etc/hoge")
		if a != "/tmp/etc/hoge" {
			t.Fatalf("%s should be /tmp/etc/hoge", a)
		}
	}
}
