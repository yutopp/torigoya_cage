//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import (
	_ "testing"
	_ "net"
	_ "os"
	_ "fmt"
	_ "path/filepath"
	_ "time"
	_ "strconv"

	_ "github.com/ugorji/go/codec"
)


/*
func TestProtocolServer(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "test_proc_profiles"))
	if err != nil {
		t.Fatalf(err.Error())
	}

	e := make(chan error)
	go RunServer("", 12321, ctx, e)
	if err := <- e; err != nil {
		t.Fatal(err)
	}

	conn, err := net.Dial("tcp", ":12321")
	if err != nil {
		t.Fatal(err)
		// handle error
	}



	//
	base_name := "aaa6" + strconv.FormatInt(time.Now().Unix(), 10)

	//
	sources := []*SourceData{
		&SourceData{
			"prog.cpp",
			[]byte(`
#include <iostream>

int main() {
	std::cout << "hello!" << std::endl;
	int i;
	std::cin >> i;
	std::cout << "input is " << i << std::endl;
}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				stdin: nil,
				setting: &ExecutionSetting{
					CpuTimeLimit: 10,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},

			Input{
				stdin: &SourceData{
					"hoge.in",
					[]byte("100"),
					false,
				},
				setting: &ExecutionSetting{
					CpuTimeLimit: 10,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},
		},
	}

	//
	ticket := &Ticket{
		BaseName: base_name,
		ProcId: 0,
		ProcVersion: "test",
		Sources: sources,
		BuildInst: build_inst,
		RunInst: run_inst,
	}





	//
	var handler ProtocolHandler

	// request
	if err := handler.WriteRequest(conn, ticket); err != nil {
		t.Fatalf("server recv: %v\n", err)
	}


	//
	for {
		kind, data, err := handler.read(conn)
		if err != nil {
			t.Fatalf("client error: %v\n", err)
			break
		}

		fmt.Printf("client recv: %d / %v\n", kind, data)
	}
}
*/
