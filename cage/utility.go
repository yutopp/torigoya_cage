//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

// +build linux

package torigoya

import (
	"os"
	"path/filepath"
)

func NormalizePath(baseDir string, path string) string {
	if filepath.IsAbs(path) {
		return path
	} else {
		return filepath.Join(path, path)
	}
}

func expectRoot() {
	if os.Geteuid() != 0 {
		panic("run this program as root")
	}
}

func readUInt(v interface{}) (uint64, bool) {
	switch v.(type) {
	case int64:
		return uint64(v.(int64)), true
	case uint64:
		return v.(uint64), true
	default:
		return 0, false
	}
}
