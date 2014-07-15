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
	"os"
	"path/filepath"
)

func TestUnitProfileStructure(t *testing.T) {
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

	profile, err := makeProcProfileFromBufAsJSON([]byte(file))
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


func TestUnitProcIndexListStructure(t *testing.T) {
	file := `
# languages

-
  id: 0
  name: "C++"
  runnable: true
  path: "lang.c++.test"

-
  id: 10
  name: "Hoge"
  runnable: false
  path: "lang.hoge.test"
`

	index_list, err := makeProcDescriptionListFromBuf([]byte(file))
	if err != nil {
		t.Fatalf("error: %v", err)
		return
	}


	if len(index_list) != 2 {
		t.Fatalf("length of index_list should be 2(but %v)", len(index_list))
	}


	if index_list[0].Id != 0 {
		t.Fatalf("index_list[0].Id should be 0(but %v)", index_list[0].Id)
	}
	if index_list[0].Name != "C++" {
		t.Fatalf("index_list[0].Name should be C++(but %v)", index_list[0].Name)
	}
	if index_list[0].Runnable != true {
		t.Fatalf("index_list[0].Runnable should be true(but %v)", index_list[0].Runnable)
	}
	if index_list[0].Path != "lang.c++.test" {
		t.Fatalf("index_list[0].Path should be lang.c++.test(but %v)", index_list[0].Path)
	}


	if index_list[1].Id != 10 {
		t.Fatalf("index_list[1].Id should be 10(but %v)", index_list[1].Id)
	}
	if index_list[1].Name != "Hoge" {
		t.Fatalf("index_list[1].Name should be Hoge(but %v)", index_list[1].Name)
	}
	if index_list[1].Runnable != false {
		t.Fatalf("index_list[1].Runnable should be false(but %v)", index_list[1].Runnable)
	}
	if index_list[1].Path != "lang.hoge.test" {
		t.Fatalf("index_list[1].Path should be lang.hoge.test(but %v)", index_list[1].Path)
	}
}

func TestLoadProcProfilesFromFile(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	index_list, err := makeProcDescriptionListFromPath(filepath.Join(gopath, "files", "proc_profiles_for_core_test", "languages.yml"))
	if err != nil {
		t.Fatalf("error: %v", err)
		return
	}


	if len(index_list) != 1 {
		t.Fatalf("length of index_list should be 1(but %v)", len(index_list))
	}


	if index_list[0].Id != 0 {
		t.Fatalf("index_list[0].Id should be 0(but %v)", index_list[0].Id)
	}
	if index_list[0].Name != "C++" {
		t.Fatalf("index_list[0].Name should be C++(but %v)", index_list[0].Name)
	}
	if index_list[0].Runnable != true {
		t.Fatalf("index_list[0].Runnable should be true(but %v)", index_list[0].Runnable)
	}
	if index_list[0].Path != "lang.c++.test" {
		t.Fatalf("index_list[0].Path should be lang.c++.test(but %v)", index_list[0].Path)
	}
}


func TestLoadProcConfigs(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	configs, err := LoadProcConfigs(filepath.Join(gopath, "files", "proc_profiles_for_core_test"))
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	_ = configs
}


func TestCmdLine(t *testing.T) {
	pd := PhaseDetail{
		Command: "g++",
		AllowedCommandLine: map[string]SelectableCommand{
			"-std=": SelectableCommand{
				Default: []string{ "c++11" },
				Select: []string{ "c++11", "c++1y" },
			},
			"-ftemplate-depth=": SelectableCommand{
				Select: []string{ "512", "1024", "2048", "4096" },
			},
			"-E": SelectableCommand{},
		},
		FixedCommandLine: [][]string{
			[]string{ "-c", "prog.cpp" },
			[]string{ "-o", "prog.o" },
		},
	}

	cmd, err := pd.MakeCompleteArgs("hogefuga \"foo bar\" -c=2", [][]string{[]string{"-std=", "c++1y"}, []string{"-E"}})
	if err != nil {
		t.Fatalf(err.Error())
	}

	var expected = []string{"g++", "-std=", "c++1y", "-E", "-c", "prog.cpp", "-o", "prog.o", "hogefuga", "foo bar", "-c=2"}
	if len(cmd) != len(expected) { t.Fatalf("expected %v (but %v)", expected, cmd) }
	for i, _ := range expected {
		if cmd[i] != expected[i] { t.Fatalf("expected %v (but %v)", expected, cmd) }
	}
}
