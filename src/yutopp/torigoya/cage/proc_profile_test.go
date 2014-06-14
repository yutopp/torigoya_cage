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
	"log"

	"gopkg.in/v1/yaml"
)

type A struct {
	Default		[]string
	Select		[]string `yaml:"select,flow"`
}

type H struct {
	File					string
	Extension				string
	Command					string
	Env						map[string]string
	AllowedCommandLine	map[string]A `yaml:"allowed_command_line"`
}

type Config struct {
	Version					string
	Is_build_required		bool
	Is_link_independent		bool

	Source					H
	Compile					H
	Link					H
	Run						 H
}

func TestA(t *testing.T) {
	file := `
---
version: '3.4'
is_build_required: true
is_link_independent: true
source:
  file: prog
  extension: cpp
compile:
  file: prog
  extension: o
  command: clang++
  env:
    PATH: /usr/local/torigoya/clang-3.4/bin:/usr/bin
    CPATH: /usr/local/torigoya/libc++-trunk/include/c++/v1
  allowed_command_line:
    -std=:
      default: c++11
      select:
      - c++1y
      - gnu++1y
      - c++11
      - gnu++11
      - c++98
      - gnu++98
    -ftemplate-depth=:
      select:
      - '512'
      - '1024'
      - '2048'
      - '4096'
    -O:
      default: '2'
      select:
      - '0'
      - '1'
      - '2'
      - '3'
    -W:
      default:
      - all
      - extra
      select:
      - all
      - extra
    -E:
    -P:
    -I:
      select:
      - /usr/local/torigoya/boost-1.55.0/include
      - /usr/local/torigoya/sprout-trunk/include
      - /usr/local/torigoya/boost-1.54.0/include
  fixed_command_line:
  - '-c ': prog.cpp
  - '-o ': prog.o
link:
  file: prog
  extension: out
  command: clang++
  env:
    PATH: /usr/local/torigoya/clang-3.4/bin:/usr/bin
    LD_LIBRARY_PATH: /usr/local/torigoya/libc++-trunk/lib
    CPATH: /usr/local/torigoya/libc++-trunk/include/c++/v1
  fixed_command_line:
  - ' ': prog.o
  - '-o ': prog.out
  - -stdlib=: libc++
  - -L: /usr/local/torigoya/libc++-trunk/lib
  - -l: pthread
run:
  command: ./prog.out
  env:
    LD_LIBRARY_PATH: /usr/local/torigoya/libc++-trunk/lib
  allowed_command_line:
  fixed_command_line:
`

	config := Config{}

	if err := yaml.Unmarshal([]byte(file), &config); err != nil {
		log.Fatalf("error: %v", err)
		return
	}
	log.Fatalf("--- t:\n%v\n\n", config)
}
