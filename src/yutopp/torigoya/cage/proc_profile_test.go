//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import (
	"testing"
	_ "log"
)

func TestA(t *testing.T) {
	file := `
---
version: HEAD-2014.4.9.e912167e7ecf
is_build_required: true
is_link_independent: false
source:
  file: Prog
  extension: java
compile:
  file: Prog
  extension: class
  command: javac
  env:
    PATH: "/usr/local/torigoya/java9-trunk/bin:/usr/bin"
  allowed_command_line:
  fixed_command_line:
  - - " "
    - Prog.java
  - - "-J-Xms"
    - 64m
  - - "-J-Xmx"
    - 128m
  - - "-J-Xss"
    - 512k
  - - "-J-XX:CompressedClassSpaceSize="
    - 32M
  - - "-J-XX:MaxMetaspaceSize="
    - 128M
  - - "-J-XX:MetaspaceSize="
    - 64M
run:
  command: java
  env:
    PATH: "/usr/local/torigoya/java9-trunk/bin:/usr/bin"
  allowed_command_line:
  fixed_command_line:
  - - "-Xms"
    - 64m
  - - "-Xmx"
    - 128m
  - - "-Xss"
    - 512k
  - - "-XX:CompressedClassSpaceSize="
    - 32M
  - - "-XX:MaxMetaspaceSize="
    - 128M
  - - "-XX:MetaspaceSize="
    - 64M
  - - " "
    - Prog
`

	profile, err := MakeProcProfile([]byte(file))
	if err != nil {
		t.Fatalf("error: %v", err)
		return
	}

	if profile.Version != "HEAD-2014.4.9.e912167e7ecf" {
		t.Fatalf("profile.Version should be HEAD-2014.4.9.e912167e7ecf(but %v)", profile.Version)
	}


	if profile.IsBuildRequired != true {
		t.Fatalf("profile.IsBuildRequired should be true(but %v)", profile.IsBuildRequired)
	}


	if profile.IsLinkIndependent != false {
		t.Fatalf("profile.IsLinkIndependent should be false(but %v)", profile.IsLinkIndependent)
	}


	//log.Fatalf("--- t:\n%v\n\n", profile)
}
