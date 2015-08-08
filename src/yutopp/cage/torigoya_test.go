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
	"fmt"
	"bytes"
	"sync"
	"strconv"
	"errors"
	"path/filepath"
)



/*
func TestCreateTarget(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	base_name := "aaa" + strconv.FormatInt(time.Now().Unix(), 10)
	content := &TextContent{
		"prog.cpp",
		[]byte("test test test"),
	}
	user_id := 1000
	group_id := 1000

	source_full_path, err := ctx.createTarget(base_name, user_id, group_id, content)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	t.Logf("source path: %s", source_full_path)
	if source_full_path != filepath.Join(ctx.sandboxDir, base_name, ctx.jailedUserDir, content.Name) {
		t.Errorf(fmt.Sprintf("%s", source_full_path))
		return
	}

	//file, err := os.Open(source_full_path)
}


func TestCreateTargetRepeat(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	base_name := "aaa" + strconv.FormatInt(time.Now().Unix(), 10)
	content := &TextContent{
		"prog.cpp",
		[]byte("test test test"),
	}
	user_id := 1000
	group_id := 1000

	for i:=0; i<2; i++ {
		if _, err := ctx.createTarget(base_name, user_id, group_id, content); err != nil {
			t.Fatalf(err.Error())
		}
	}
}







func TestCreateInput(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	base_name := "aaa3" + strconv.FormatInt(time.Now().Unix(), 10)
	content := &TextContent{
		"prog.cpp",
		[]byte("test test test"),
	}
	user_id := 1000
	group_id := 1000

	stdin := &TextContent{
		"in" + strconv.FormatInt(time.Now().Unix(), 10),
		[]byte("iniini~~~"),
	}

	ctx.createTarget(base_name, user_id, group_id, content)

	user_dir_path, input_path, err := ctx.reassignTarget(base_name, user_id, group_id, func(base_directory_name string) (*string, error) {
		path, err := ctx.createInput(base_directory_name, group_id, stdin)
		if err != nil { return nil, err }
		return &path, nil
	})
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	t.Logf("user dir path: %s", user_dir_path)
	if user_dir_path != filepath.Join(ctx.sandboxDir, base_name) {
		t.Errorf(fmt.Sprintf("%s", user_dir_path))
		return
	}

	if input_path == nil {
		t.Errorf("length of input_paths should not be nil (%v)", input_path)
		return
	}

	//
	t.Logf("input path: %s", *input_path)
	if *input_path != filepath.Join(ctx.sandboxDir, base_name, ctx.jailedUserDir, "stdin", stdin.Name) {
		t.Errorf("%v", *input_path)
		return
	}

}

func TestInvokeBuild(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	base_name := "aaa5" + strconv.FormatInt(time.Now().Unix(), 10)

	//
	sources := []*SourceData{
		&SourceData{
			"prog.cpp",
			[]byte(`
#include <iostream>

int main() {
	std::cout << "hello!" << std::endl;
}
`),
			false,
		},
	}

	// load id:0/version:0.0.0
	configs, _ := LoadProcConfigs(filepath.Join(gopath, "files", "proc_profiles_for_core_test"))
	proc_profile := configs[0].Versioned["test"]

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

	// build
	if err := ctx.execManagedBuild(&proc_profile, base_name, sources, build_inst, nil); err != nil {
		t.Errorf(err.Error())
		return
	}
}

*/


func TestTicketBasicUnit(t *testing.T) {
	ctx, err := InitContext(makeDefaultCtxOpt())
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	sources := []*SourceData{
		&SourceData{
			"prog.c",
			[]byte(`
#include <stdio.h>

int main() {
	printf("hello!\n");
	int i;
	scanf("%d", &i);
	printf("input is %d\n", i);

	return 0;
}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/gcc", "prog.c", "-c", "-o", "prog.o"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/gcc", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				Stdin: nil,
				RunSetting: &ExecutionSetting{
					Args: []string{"./prog.out"},
					Envs: []string{},
					CpuTimeLimit: 10,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},

			Input{
				Stdin: &SourceData{
					"hoge.in",
					[]byte("100"),
					false,
				},
				RunSetting: &ExecutionSetting{
					Args: []string{"./prog.out"},
					Envs: []string{},
					CpuTimeLimit: 10,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},
		},
	}

	//
	ticket := &Ticket{
		BaseName: "",
		Sources: sources,
		BuildInst: build_inst,
		RunInst: run_inst,
	}

	// execute
	var result test_result
	result.run = make(map[int]*test_result_unit)
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	t.Logf("%V", result)

	//
	expect_result := testExpectResult{
		compile: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		link: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		run: map[int]*testExpectResultUnit{
			0: &testExpectResultUnit{
				out: []byte("hello!\ninput is 0\n"),
				status: &testExpectStatus{
					exited: BoolOpt(true),
					exitStatus: IntOpt(0),
				},
			},
			1: &testExpectResultUnit{
				out: []byte("hello!\ninput is 100\n"),
				status: &testExpectStatus{
					exited: BoolOpt(true),
					exitStatus: IntOpt(0),
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}

func TestTicketMultiSource(t *testing.T) {
	ctx, err := InitContext(makeDefaultCtxOpt())
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	sources := []*SourceData{
		&SourceData{
			"hoge.hpp",
			[]byte(`
#include <iostream>

namespace hoge {
	void foo() {
		std::cout << "foo" << std::endl;
	}
}
`),
			false,
		},

		&SourceData{
			"prog.cpp",
			[]byte(`
#include "hoge.hpp"

int main() {
	hoge::foo();
}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs: []string{},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				Stdin: nil,
				RunSetting: &ExecutionSetting{
					Args: []string{"./prog.out"},
					Envs: []string{},
					CpuTimeLimit: 10,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},
		},
	}

	//
	ticket := &Ticket{
		BaseName: "",
		Sources: sources,
		BuildInst: build_inst,
		RunInst: run_inst,
	}

	// execute
	var result test_result
	result.run = make(map[int]*test_result_unit)
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	t.Logf("%V", result)

	//
	expect_result := testExpectResult{
		compile: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		link: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		run: map[int]*testExpectResultUnit{
			0: &testExpectResultUnit{
				out: []byte("foo\n"),
				status: &testExpectStatus{
					exited: BoolOpt(true),
					exitStatus: IntOpt(0),
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}

/*
func TestTicketPS(t *testing.T) {
	ctx, err := InitContext(makeDefaultCtxOpt())
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	sources := []*SourceData{
		&SourceData{
			"prog.cpp",
			[]byte(`
#include <cstdlib>

int main() {
	std::system("ps aux");
}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs: []string{},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				Stdin: nil,
				RunSetting: &ExecutionSetting{
					Args: []string{"./prog.out"},
					Envs: []string{},
					CpuTimeLimit: 10,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},
		},
	}

	//
	ticket := &Ticket{
		BaseName: "",
		Sources: sources,
		BuildInst: build_inst,
		RunInst: run_inst,
	}

	// execute
	var result test_result
	result.run = make(map[int]*test_result_unit)
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	expect_result := testExpectResult{
		compile: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		link: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		run: map[int]*testExpectResultUnit{
			0: &testExpectResultUnit{
				out: []byte("hello!\ninput is 0\n"),
				status: &testExpectStatus{
					exited: BoolOpt(true),
					exitStatus: IntOpt(0),
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}
*/

func TestTicketSignal(t *testing.T) {
	ctx, err := InitContext(makeDefaultCtxOpt())
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	sources := []*SourceData{
		&SourceData{
			"prog.cpp",
			[]byte(`
#include <stdio.h>
#include <signal.h>
#include <errno.h>
#include <string.h>

int main() {
    puts("hello!");
	fflush(stdout);
    if ( raise(9) != 0 ) {
        printf("errno=%d : %s\\n", errno, strerror( errno ));
    }
    puts("unreachable!");

    return 0;
}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs: []string{},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				Stdin: nil,
				RunSetting: &ExecutionSetting{
					Args: []string{"./prog.out"},
					Envs: []string{},
					CpuTimeLimit: 10,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},
		},
	}

	//
	ticket := &Ticket{
		BaseName: "",
		Sources: sources,
		BuildInst: build_inst,
		RunInst: run_inst,
	}

	// execute
	var result test_result
	result.run = make(map[int]*test_result_unit)
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	t.Logf("%V", result)

	//
	expect_result := testExpectResult{
		compile: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		link: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		run: map[int]*testExpectResultUnit{
			0: &testExpectResultUnit{
				out: []byte("hello!\n"),
				status: &testExpectStatus{
					exited: BoolOpt(false),
					signaled: BoolOpt(true),
					signal: IntOpt(9),
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}

func TestTicketBasicParallel1(t *testing.T) {
	ctx, err := InitContext(makeDefaultCtxOpt())
	if err != nil {
		t.Errorf(err.Error())
		return
	}

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
			Args: []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs: []string{},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				Stdin: nil,
				RunSetting: &ExecutionSetting{
					Args: []string{"./prog.out"},
					Envs: []string{},
					CpuTimeLimit: 10,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},

			Input{
				Stdin: &SourceData{
					"hoge.in",
					[]byte("100"),
					false,
				},
				RunSetting: &ExecutionSetting{
					Args: []string{"./prog.out"},
					Envs: []string{},
					CpuTimeLimit: 10,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},
		},
	}

	//
	ticket := &Ticket{
		BaseName: "",
		Sources: sources,
		BuildInst: build_inst,
		RunInst: run_inst,
	}

	//
	expect_result := testExpectResult{
		compile: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		link: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		run: map[int]*testExpectResultUnit{
			0: &testExpectResultUnit{
				out: []byte("hello!\ninput is 0\n"),
				status: &testExpectStatus{
					exited: BoolOpt(true),
					exitStatus: IntOpt(0),
				},
			},
			1: &testExpectResultUnit{
				out: []byte("hello!\ninput is 100\n"),
				status: &testExpectStatus{
					exited: BoolOpt(true),
					exitStatus: IntOpt(0),
				},
			},
		},
	}

	// run in parallel
	wg := new(sync.WaitGroup)
	m := new(sync.Mutex)
	const num = 16
	var fx [num]bool
	for i := 0; i < num; i++ {
		wg.Add(1)

		go func(no int) {
			defer func() {
				fmt.Printf("Done! %d\n", no)

				fmt.Printf("fs! %v\n", fx)
				wg.Done()
			}()

			// execute
			var result test_result
			result.run = make(map[int]*test_result_unit)
			f := makeHelperCallback(&result)
			if err := ctx.ExecTicket(ticket, f); err != nil {
				t.Errorf(err.Error())
				return
			}

			t.Logf("%V", result)
			assertTestResult(t, &result, &expect_result)

			m.Lock()
			fx[no] = true	// succeeded
			m.Unlock()
		}(i)
	}

	wg.Wait()
}

func TestTicketBasicParallel2(t *testing.T) {
	ctx, err := InitContext(makeDefaultCtxOpt())
	if err != nil {
		t.Errorf(err.Error())
		return
	}

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
			Args: []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs: []string{},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}


	// run in parallel
	wg := new(sync.WaitGroup)
	m := new(sync.Mutex)
	const num = 30	// More!
	var fx [num]bool
	for i := 0; i < num; i++ {
		wg.Add(1)

		go func(no int) {
			defer func() {
				fmt.Printf("Done! %d\n", no)

				fmt.Printf("fs! %v\n", fx)
				wg.Done()
			}()

			//
			run_inst := &RunInstruction{
				Inputs: []Input{
					Input{
						Stdin: &SourceData{
							Data: []byte(strconv.Itoa(no)),
						},
						RunSetting: &ExecutionSetting{
							Args: []string{"./prog.out"},
							Envs: []string{},
							CpuTimeLimit: 10,
							MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
						},
					},
				},
			}

			//
			ticket := &Ticket{
				BaseName: "",
				Sources: sources,
				BuildInst: build_inst,
				RunInst: run_inst,
			}

			//
			expect_result := testExpectResult{
				compile: testExpectResultUnit{
					status: &testExpectStatus{
						exited: BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				link: testExpectResultUnit{
					status: &testExpectStatus{
						exited: BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				run: map[int]*testExpectResultUnit{
					0: &testExpectResultUnit{
						out: []byte(fmt.Sprintf("hello!\ninput is %d\n", no)),
						status: &testExpectStatus{
							exited: BoolOpt(true),
							exitStatus: IntOpt(0),
						},
					},
				},
			}

			// execute
			var result test_result
			result.run = make(map[int]*test_result_unit)
			f := makeHelperCallback(&result)
			if err := ctx.ExecTicket(ticket, f); err != nil {
				t.Errorf(err.Error())
				return
			}

			t.Logf("%V", result)
			assertTestResult(t, &result, &expect_result)

			m.Lock()
			fx[no] = true	// succeeded
			m.Unlock()
		}(i)
	}

	wg.Wait()
}

func TestTicketTLE(t *testing.T) {
	ctx, err := InitContext(makeDefaultCtxOpt())
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	sources := []*SourceData{
		&SourceData{
			"prog.cpp",
			[]byte(`
#include <iostream>

int main() {
	for(;;);
}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs: []string{},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				Stdin: nil,
				RunSetting: &ExecutionSetting{
					Args: []string{"./prog.out"},
					Envs: []string{},
					CpuTimeLimit: 1,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},
		},
	}

	//
	ticket := &Ticket{
		BaseName: "",
		Sources: sources,
		BuildInst: build_inst,
		RunInst: run_inst,
	}

	// execute
	var result test_result
	result.run = make(map[int]*test_result_unit)
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	t.Logf("%V", result)

	//
	expect_result := testExpectResult{
		compile: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		link: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		run: map[int]*testExpectResultUnit{
			0: &testExpectResultUnit{
				status: &testExpectStatus{
					exited: BoolOpt(false),		// killed
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}

func TestTicketSleepTLE(t *testing.T) {
	ctx, err := InitContext(makeDefaultCtxOpt())
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	sources := []*SourceData{
		&SourceData{
			"prog.cpp",
			[]byte(`
#include <unistd.h>

int main() {
	::sleep(256);
}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs: []string{},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				Stdin: nil,
				RunSetting: &ExecutionSetting{
					Args: []string{"./prog.out"},
					Envs: []string{},
					CpuTimeLimit: 1,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},
		},
	}

	//
	ticket := &Ticket{
		BaseName: "",
		Sources: sources,
		BuildInst: build_inst,
		RunInst: run_inst,
	}

	// execute
	var result test_result
	result.run = make(map[int]*test_result_unit)
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	t.Logf("%V", result)

	//
	expect_result := testExpectResult{
		compile: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		link: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		run: map[int]*testExpectResultUnit{
			0: &testExpectResultUnit{
				status: &testExpectStatus{
					exited: BoolOpt(false),		// killed
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}

func TestTicketMLE(t *testing.T) {
	ctx, err := InitContext(makeDefaultCtxOpt())
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	sources := []*SourceData{
		&SourceData{
			"prog.cpp",
			[]byte(`
#include <iostream>

int main() {
	std::size_t s;
	std::cin >> s;

	char* buffer = new char[s]{};
	std::cout << "allocated: " << s << std::endl;
	buffer[s-1] = 'A';
}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs: []string{},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				Stdin: &SourceData{
					Data: []byte("300000000"),	// 300MB
				},
				RunSetting: &ExecutionSetting{
					Args: []string{"./prog.out"},
					Envs: []string{},
					CpuTimeLimit: 1,
					MemoryBytesLimit: 200 * 1024 * 1024,	// 200MB
				},
			},
		},
	}

	//
	ticket := &Ticket{
		BaseName: "",
		Sources: sources,
		BuildInst: build_inst,
		RunInst: run_inst,
	}

	// execute
	var result test_result
	result.run = make(map[int]*test_result_unit)
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	t.Logf("%V", result)

	//
	expect_result := testExpectResult{
		compile: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		link: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		run: map[int]*testExpectResultUnit{
			0: &testExpectResultUnit{
				status: &testExpectStatus{
					exited: BoolOpt(false),		// killed
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}

func TestTicketRepeat(t *testing.T) {
	ctx, err := InitContext(makeDefaultCtxOpt())
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	sources := []*SourceData{
		&SourceData{
			"prog.cpp",
			[]byte(`
#include <iostream>

int main() {
	for(int i=0; i<100000; ++i) {
		std::cout << i << "\n" << std::flush;
	}
}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs: []string{},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit: 10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				Stdin: nil,
				RunSetting: &ExecutionSetting{
					Args: []string{"./prog.out"},
					Envs: []string{},
					CpuTimeLimit: 10,
					MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
				},
			},
		},
	}

	//
	ticket := &Ticket{
		BaseName: "",
		Sources: sources,
		BuildInst: build_inst,
		RunInst: run_inst,
	}

	// execute
	var result test_result
	result.run = make(map[int]*test_result_unit)
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	expect_out := []byte{}
	for i:=0; i<100000; i++ {
		expect_out = append(expect_out, fmt.Sprintf("%d\n", i)...)
	}

	//
	expect_result := testExpectResult{
		compile: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		link: testExpectResultUnit{
			status: &testExpectStatus{
				exited: BoolOpt(true),
				exitStatus: IntOpt(0),
			},
		},
		run: map[int]*testExpectResultUnit{
			0: &testExpectResultUnit{
				out: expect_out,
				status: &testExpectStatus{
					exited: BoolOpt(true),
					exitStatus: IntOpt(0),
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}


// ==================================================
// ==================================================
//
func makeDefaultCtxOpt() *ContextOptions {
	gopath := os.Getenv("GOPATH")

	executor := &awahoSandboxExecutor{
		ExecutablePath: filepath.Join(gopath, "_awaho/awaho"),
	}

	return &ContextOptions{
		BasePath: gopath,
		UserFilesBasePath: "/tmp/cage_test",
		PackageInstalledBasePath: "/usr/local/procgarden",

		SandboxExec: executor,
	}
}


// ==================================================
// ==================================================
//
func assertTestResult(t *testing.T, result *test_result, expect *testExpectResult) {
	assertUnit := func (tag string, result *test_result_unit, expect *testExpectResultUnit) {
		// check status(a.k.a result)
		if expect.status != nil {
			// expected is specified
			if result.result == nil {
				t.Fatalf("[ERROR  : %s / result] result is nil", tag)
			}

			if expect.status.exited.Exists {
				if expect.status.exited.Value != result.result.Exited {
					t.Fatalf("ERROR  : [%s / result.Exited] Expect(%s) but returned(%s)", tag,
						expect.status.exited.Value,
						result.result.Exited,
					)
				}
			} else {
				t.Logf("[SKIPPED: %s / result.Exited]", tag)
			}

			if expect.status.exitStatus.Exists {
				if expect.status.exitStatus.Value != result.result.ExitStatus {
					t.Fatalf("ERROR  : [%s / result.ExitStatus] Expect(%s) but returned(%s)", tag,
						expect.status.exitStatus.Value,
						result.result.ExitStatus,
					)
				}
			} else {
				t.Logf("[SKIPPED: %s / result.ExitStatus]", tag)
			}

			if expect.status.signaled.Exists {
				if expect.status.signaled.Value != result.result.Signaled {
					t.Fatalf("ERROR  : [%s / result.Exited] Expect(%s) but returned(%s)", tag,
						expect.status.signaled.Value,
						result.result.Signaled,
					)
				}
			} else {
				t.Logf("[SKIPPED: %s / result.Signaled]", tag)
			}

			if expect.status.signal.Exists {
				if expect.status.signal.Value != result.result.Signal {
					t.Fatalf("ERROR  : [%s / result.Exited] Expect(%s) but returned(%s)", tag,
						expect.status.signal.Value,
						result.result.Signal,
					)
				}
			} else {
				t.Logf("[SKIPPED: %s / result.Signal]", tag)
			}

		} else {
			t.Logf("[SKIPPED: %s / result]", tag)
		}

		if expect.out != nil {
			if result.out == nil {
				t.Fatalf("[ERROR  : %s / out] result is nil", tag)
			}
			if !bytes.Equal(expect.out, result.out) {
				t.Fatalf("[ERROR  : %s / out] Expect(%s) but returned(%s)", tag, expect.out, result.out)
			}

		} else {
			t.Logf("[SKIPPED: %s / out]", tag)
		}

		if expect.err != nil {
			if result.out == nil {
				t.Fatalf("[ERROR  : %s / err] result is nil", tag)
			}
			if !bytes.Equal(expect.err, result.err) {
				t.Fatalf("[ERROR  : %s / err] Expect(%s) but returned(%s)", tag, expect.err, result.err)
			}

		} else {
			t.Logf("[SKIPPED: %s / err]", tag)
		}
	}

	assertUnit("compile", &result.compile, &expect.compile)
	assertUnit("link   ", &result.link, &expect.link)

	// run
	checked := make(map[int]bool)
	for key, result_unit := range result.run {
		expect_unit, ok := expect.run[key]
		if !ok {
			t.Fatalf("Unexpected key(%d)", key)
		}
		assertUnit(fmt.Sprintf("run:%d", key), result_unit, expect_unit)
		checked[key] = true
	}

	for key, _ := range expect.run {
		if _, ok := checked[key]; !ok {
			t.Fatalf("The key(%d) was not checked", key)
		}
	}
}


type testExpectResult struct {
	compile, link	testExpectResultUnit
	run				map[int]*testExpectResultUnit
}
type testExpectResultUnit struct {
	out, err	[]byte
	status		*testExpectStatus
}
type testExpectStatus struct {
	exited				BoolOptionalType
	exitStatus			IntOptionalType
	signaled			BoolOptionalType
	signal				IntOptionalType
}



type test_result_unit struct {
	out, err	[]byte
	result		*ExecutedResult
}
type test_result struct {
	compile, link	test_result_unit
	run				map[int]*test_result_unit
}

func makeHelperCallback(result *test_result) func(v interface{}) error {
	return func(v interface{}) error {
		switch v.(type) {
		case *StreamExecutedResult:
			r := v.(*StreamExecutedResult)
			switch r.Mode {
			case CompileMode:
				result.compile.result = r.Result
			case LinkMode:
				result.link.result = r.Result
			case RunMode:
				if result.run[r.Index] == nil { result.run[r.Index] = &test_result_unit{} }
				result.run[r.Index].result = r.Result
			default:
				return errors.New("unsupported mode.")
			}

		case *StreamOutputResult:
			r := v.(*StreamOutputResult)
			switch r.Mode {
			case CompileMode:
				switch r.Output.Fd {
				case StdoutFd:
					result.compile.out = append(result.compile.out, r.Output.Buffer...)
				case StderrFd:
					result.compile.err = append(result.compile.err, r.Output.Buffer...)
				default:
					return errors.New("unsupported fd.")
				}

			case LinkMode:
				switch r.Output.Fd {
				case StdoutFd:
					result.link.out = append(result.link.out, r.Output.Buffer...)
				case StderrFd:
					result.link.err = append(result.link.err, r.Output.Buffer...)
				default:
					return errors.New("unsupported fd.")
				}

			case RunMode:
				if result.run[r.Index] == nil { result.run[r.Index] = &test_result_unit{} }
				switch r.Output.Fd {
				case StdoutFd:
					result.run[r.Index].out = append(result.run[r.Index].out, r.Output.Buffer...)
				case StderrFd:
					result.run[r.Index].err = append(result.run[r.Index].err, r.Output.Buffer...)
				default:
					return errors.New("unsupported fd.")
				}

			default:
				return errors.New("unsupported mode.")
			}

		default:
			return errors.New("unsupported type.")
		}

		return nil
	}
}


type OptionalBase struct {
	Exists	bool
}

type IntOptionalType struct {
	OptionalBase
	Value	int
}

func IntOpt(v int) IntOptionalType {
	return IntOptionalType{
		OptionalBase: OptionalBase{true},
		Value: v,
	}
}

type BoolOptionalType struct {
	OptionalBase
	Value	bool
}

func BoolOpt(v bool) BoolOptionalType {
	return BoolOptionalType{
		OptionalBase: OptionalBase{true},
		Value: v,
	}
}
