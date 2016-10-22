//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
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
	buildInst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/gcc", "prog.c", "-c", "-o", "prog.o"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/gcc", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	runInsts := []*RunInstruction{
		&RunInstruction{
			Stdin: nil,
			RunSetting: &ExecutionSetting{
				Args:             []string{"./prog.out"},
				Envs:             []string{},
				CpuTimeLimit:     10,
				MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
			},
		},
		&RunInstruction{
			Stdin: &SourceData{
				"hoge.in",
				[]byte("100"),
				false,
			},
			RunSetting: &ExecutionSetting{
				Args:             []string{"./prog.out"},
				Envs:             []string{},
				CpuTimeLimit:     10,
				MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
			},
		},
	}

	execSpecs := []*ExecutionSpec{
		&ExecutionSpec{
			BuildInst: buildInst,
			RunInsts:  runInsts,
		},
	}

	//
	ticket := &Ticket{
		BaseName:  "TestTicketBasicUnit",
		Sources:   sources,
		ExecSpecs: execSpecs,
	}

	// execute
	var result testResult
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	expectedResult := testExpectedResult{
		execResults: []testExpectedExecResult{
			testExpectedExecResult{
				compile: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				link: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				run: []testExpectedUnitResult{
					testExpectedUnitResult{
						out: []byte("hello!\ninput is 0\n"),
						status: &testExpectedStatus{
							exited:     BoolOpt(true),
							exitStatus: IntOpt(0),
						},
					},
					testExpectedUnitResult{
						out: []byte("hello!\ninput is 100\n"),
						status: &testExpectedStatus{
							exited:     BoolOpt(true),
							exitStatus: IntOpt(0),
						},
					},
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expectedResult)
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
	buildInst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args:             []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs:             []string{},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	runInsts := []*RunInstruction{
		&RunInstruction{
			Stdin: nil,
			RunSetting: &ExecutionSetting{
				Args:             []string{"./prog.out"},
				Envs:             []string{},
				CpuTimeLimit:     10,
				MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
			},
		},
	}

	execSpecs := []*ExecutionSpec{
		&ExecutionSpec{
			BuildInst: buildInst,
			RunInsts:  runInsts,
		},
	}
	//
	ticket := &Ticket{
		BaseName:  "TestTicketMultiSource",
		Sources:   sources,
		ExecSpecs: execSpecs,
	}

	// execute
	var result testResult
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	expectedResult := testExpectedResult{
		execResults: []testExpectedExecResult{
			testExpectedExecResult{
				compile: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				link: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				run: []testExpectedUnitResult{
					testExpectedUnitResult{
						out: []byte("foo\n"),
						status: &testExpectedStatus{
							exited:     BoolOpt(true),
							exitStatus: IntOpt(0),
						},
					},
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expectedResult)
}

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
	buildInst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args:             []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs:             []string{},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	runInsts := []*RunInstruction{
		&RunInstruction{
			Stdin: nil,
			RunSetting: &ExecutionSetting{
				Args:             []string{"./prog.out"},
				Envs:             []string{},
				CpuTimeLimit:     10,
				MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
			},
		},
	}

	execSpecs := []*ExecutionSpec{
		&ExecutionSpec{
			BuildInst: buildInst,
			RunInsts:  runInsts,
		},
	}

	//
	ticket := &Ticket{
		BaseName:  "TestTicketPS",
		Sources:   sources,
		ExecSpecs: execSpecs,
	}

	// execute
	var result testResult
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	expectedResult := testExpectedResult{
		execResults: []testExpectedExecResult{
			testExpectedExecResult{
				compile: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				link: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				run: []testExpectedUnitResult{
					testExpectedUnitResult{
						outFunc: func(buf []byte) error {
							// Example
							// USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
							// root         1  0.0  0.0  13308   180 ?        Sl+  10:37   0:00 d=(^o^)=b
							// _70sy9y+     2  0.0  0.0  13088  1592 ?        S+   10:37   0:00 ./prog.out
							// _70sy9y+     4  0.0  0.0  32852  2776 ?        R+   10:37   0:00 ps aux
							lines := strings.Split(string(buf), "\n")

							const line1Expected = "^root(\\s+)1"
							if !regexp.MustCompile(line1Expected).MatchString(lines[1]) {
								return fmt.Errorf("`%s` does not contain `%s`", lines[1], line1Expected)
							}

							const line2Expected = "^([_0-9a-z]+\\+)(\\s+)2"
							if !regexp.MustCompile(line2Expected).MatchString(lines[2]) {
								return fmt.Errorf("`%s` does not contain `%s`", lines[2], line2Expected)
							}

							return nil
						},
						status: &testExpectedStatus{
							exited:     BoolOpt(true),
							exitStatus: IntOpt(0),
						},
					},
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expectedResult)
}

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
	buildInst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args:             []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs:             []string{},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	runInsts := []*RunInstruction{
		&RunInstruction{
			Stdin: nil,
			RunSetting: &ExecutionSetting{
				Args:             []string{"./prog.out"},
				Envs:             []string{},
				CpuTimeLimit:     10,
				MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
			},
		},
	}

	execSpecs := []*ExecutionSpec{
		&ExecutionSpec{
			BuildInst: buildInst,
			RunInsts:  runInsts,
		},
	}

	//
	ticket := &Ticket{
		BaseName:  "TestTicketSignal",
		Sources:   sources,
		ExecSpecs: execSpecs,
	}

	// execute
	var result testResult
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	expectedResult := testExpectedResult{
		execResults: []testExpectedExecResult{
			testExpectedExecResult{
				compile: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				link: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				run: []testExpectedUnitResult{
					testExpectedUnitResult{
						out: []byte("hello!\n"),
						status: &testExpectedStatus{
							exited:   BoolOpt(false),
							signaled: BoolOpt(true),
							signal:   IntOpt(9),
						},
					},
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expectedResult)
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
	buildInst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args:             []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs:             []string{},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	runInsts := []*RunInstruction{
		&RunInstruction{
			Stdin: nil,
			RunSetting: &ExecutionSetting{
				Args:             []string{"./prog.out"},
				Envs:             []string{},
				CpuTimeLimit:     10,
				MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
			},
		},
		&RunInstruction{
			Stdin: &SourceData{
				"hoge.in",
				[]byte("100"),
				false,
			},
			RunSetting: &ExecutionSetting{
				Args:             []string{"./prog.out"},
				Envs:             []string{},
				CpuTimeLimit:     10,
				MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
			},
		},
	}

	execSpecs := []*ExecutionSpec{
		&ExecutionSpec{
			BuildInst: buildInst,
			RunInsts:  runInsts,
		},
	}

	//
	ticket := &Ticket{
		BaseName:  "TestTicketBasicParallel1",
		Sources:   sources,
		ExecSpecs: execSpecs,
	}

	//
	expectedResult := testExpectedResult{
		execResults: []testExpectedExecResult{
			testExpectedExecResult{
				compile: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				link: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				run: []testExpectedUnitResult{
					testExpectedUnitResult{
						out: []byte("hello!\ninput is 0\n"),
						status: &testExpectedStatus{
							exited:     BoolOpt(true),
							exitStatus: IntOpt(0),
						},
					},
					testExpectedUnitResult{
						out: []byte("hello!\ninput is 100\n"),
						status: &testExpectedStatus{
							exited:     BoolOpt(true),
							exitStatus: IntOpt(0),
						},
					},
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
				t.Logf("Done! %d", no)
				t.Logf("fs! %v", fx)
				wg.Done()
			}()

			ticketTmp := *ticket
			ticketTmp.BaseName = ticketTmp.BaseName + "-" + strconv.Itoa(no)

			// execute
			var result testResult
			f := makeHelperCallback(&result)
			if err := ctx.ExecTicket(&ticketTmp, f); err != nil {
				t.Errorf(err.Error())
				return
			}

			assertTestResult(t, &result, &expectedResult)

			m.Lock()
			fx[no] = true // succeeded
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
	buildInst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args:             []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs:             []string{},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	// run in parallel
	wg := new(sync.WaitGroup)
	m := new(sync.Mutex)
	const num = 30 // More!
	var fx [num]bool
	for i := 0; i < num; i++ {
		wg.Add(1)

		go func(no int) {
			defer func() {
				t.Logf("Done! %d", no)
				t.Logf("fs! %v", fx)
				wg.Done()
			}()

			//
			runInsts := []*RunInstruction{
				&RunInstruction{
					Stdin: &SourceData{
						Data: []byte(strconv.Itoa(no)),
					},
					RunSetting: &ExecutionSetting{
						Args:             []string{"./prog.out"},
						Envs:             []string{},
						CpuTimeLimit:     10,
						MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
					},
				},
			}

			execSpecs := []*ExecutionSpec{
				&ExecutionSpec{
					BuildInst: buildInst,
					RunInsts:  runInsts,
				},
			}

			//
			ticket := &Ticket{
				BaseName:  "TestTicketBasicParallel2-" + strconv.Itoa(no),
				Sources:   sources,
				ExecSpecs: execSpecs,
			}

			//
			expectedResult := testExpectedResult{
				execResults: []testExpectedExecResult{
					testExpectedExecResult{
						compile: testExpectedUnitResult{
							status: &testExpectedStatus{
								exited:     BoolOpt(true),
								exitStatus: IntOpt(0),
							},
						},
						link: testExpectedUnitResult{
							status: &testExpectedStatus{
								exited:     BoolOpt(true),
								exitStatus: IntOpt(0),
							},
						},
						run: []testExpectedUnitResult{
							testExpectedUnitResult{
								out: []byte(fmt.Sprintf("hello!\ninput is %d\n", no)),
								status: &testExpectedStatus{
									exited:     BoolOpt(true),
									exitStatus: IntOpt(0),
								},
							},
						},
					},
				},
			}

			// execute
			var result testResult
			f := makeHelperCallback(&result)
			if err := ctx.ExecTicket(ticket, f); err != nil {
				t.Errorf(err.Error())
				return
			}

			assertTestResult(t, &result, &expectedResult)

			m.Lock()
			fx[no] = true // succeeded
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
	buildInst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args:             []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs:             []string{},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	runInsts := []*RunInstruction{
		&RunInstruction{
			Stdin: nil,
			RunSetting: &ExecutionSetting{
				Args:             []string{"./prog.out"},
				Envs:             []string{},
				CpuTimeLimit:     1,
				MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
			},
		},
	}

	execSpecs := []*ExecutionSpec{
		&ExecutionSpec{
			BuildInst: buildInst,
			RunInsts:  runInsts,
		},
	}

	//
	ticket := &Ticket{
		BaseName:  "TestTicketTLE",
		Sources:   sources,
		ExecSpecs: execSpecs,
	}

	// execute
	var result testResult
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	expectedResult := testExpectedResult{
		execResults: []testExpectedExecResult{
			testExpectedExecResult{
				compile: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				link: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				run: []testExpectedUnitResult{
					testExpectedUnitResult{
						status: &testExpectedStatus{
							exited: BoolOpt(false), // killed
						},
					},
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expectedResult)
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
	buildInst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args:             []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs:             []string{},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	runInsts := []*RunInstruction{
		&RunInstruction{
			Stdin: nil,
			RunSetting: &ExecutionSetting{
				Args:             []string{"./prog.out"},
				Envs:             []string{},
				CpuTimeLimit:     1,
				MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
			},
		},
	}

	execSpecs := []*ExecutionSpec{
		&ExecutionSpec{
			BuildInst: buildInst,
			RunInsts:  runInsts,
		},
	}

	//
	ticket := &Ticket{
		BaseName:  "TestTicketSleepTLE",
		Sources:   sources,
		ExecSpecs: execSpecs,
	}

	// execute
	var result testResult
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	t.Logf("%V", result)

	//
	expectedResult := testExpectedResult{
		execResults: []testExpectedExecResult{
			testExpectedExecResult{
				compile: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				link: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				run: []testExpectedUnitResult{
					testExpectedUnitResult{
						status: &testExpectedStatus{
							exited: BoolOpt(false), // killed
						},
					},
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expectedResult)
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
	buildInst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args:             []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs:             []string{},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	runInsts := []*RunInstruction{
		&RunInstruction{
			Stdin: &SourceData{
				Data: []byte("300000000"), // 300MB
			},
			RunSetting: &ExecutionSetting{
				Args:             []string{"./prog.out"},
				Envs:             []string{},
				CpuTimeLimit:     1,
				MemoryBytesLimit: 200 * 1024 * 1024, // 200MB
			},
		},
	}

	execSpecs := []*ExecutionSpec{
		&ExecutionSpec{
			BuildInst: buildInst,
			RunInsts:  runInsts,
		},
	}

	//
	ticket := &Ticket{
		BaseName:  "TestTicketMLE",
		Sources:   sources,
		ExecSpecs: execSpecs,
	}

	// execute
	var result testResult
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	expectedResult := testExpectedResult{
		execResults: []testExpectedExecResult{
			testExpectedExecResult{
				compile: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				link: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				run: []testExpectedUnitResult{
					testExpectedUnitResult{
						status: &testExpectedStatus{
							exited: BoolOpt(false), // killed
						},
					},
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expectedResult)
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
	buildInst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			Args:             []string{"/usr/bin/g++", "prog.cpp", "-c", "-o", "prog.o"},
			Envs:             []string{},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			Args: []string{"/usr/bin/g++", "prog.o", "-o", "prog.out"},
			Envs: []string{
				"PATH=/usr/bin",
			},
			CpuTimeLimit:     10,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	runInsts := []*RunInstruction{
		&RunInstruction{
			Stdin: nil,
			RunSetting: &ExecutionSetting{
				Args:             []string{"./prog.out"},
				Envs:             []string{},
				CpuTimeLimit:     10,
				MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
			},
		},
	}

	execSpecs := []*ExecutionSpec{
		&ExecutionSpec{
			BuildInst: buildInst,
			RunInsts:  runInsts,
		},
	}

	//
	ticket := &Ticket{
		BaseName:  "TestTicketRepeat",
		Sources:   sources,
		ExecSpecs: execSpecs,
	}

	// execute
	var result testResult
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	expectedOut := []byte{}
	for i := 0; i < 100000; i++ {
		expectedOut = append(expectedOut, fmt.Sprintf("%d\n", i)...)
	}

	//
	expectedResult := testExpectedResult{
		execResults: []testExpectedExecResult{
			testExpectedExecResult{
				compile: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				link: testExpectedUnitResult{
					status: &testExpectedStatus{
						exited:     BoolOpt(true),
						exitStatus: IntOpt(0),
					},
				},
				run: []testExpectedUnitResult{
					testExpectedUnitResult{
						out: expectedOut,
						status: &testExpectedStatus{
							exited:     BoolOpt(true),
							exitStatus: IntOpt(0),
						},
					},
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expectedResult)
}

// ==================================================
// ==================================================
//
func makeDefaultCtxOpt() *ContextOptions {
	baseDir := os.Getenv("TORIGOYA_TEST_BASE_DIR")

	executor := &awahoSandboxExecutor{
		ExecutablePath: filepath.Join(baseDir, "_awaho/awaho"),
		HostMountDir:   filepath.Join(baseDir, "_env_test"),
		GuestMountDir:  "/usr/local/procgarden",
	}

	return &ContextOptions{
		BasePath:          baseDir,
		UserFilesBasePath: "/tmp/cage_test",
		SandboxExec:       executor,
	}
}

// ==================================================
// ==================================================
//
func assertTestResult(t *testing.T, result *testResult, expect *testExpectedResult) {
	assertUnit := func(tag string, result *testUnitResult, expect *testExpectedUnitResult) {
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
			if expect.outFunc != nil {
				if err := expect.outFunc(result.out); err != nil {
					t.Fatalf("[ERROR  : %s / out] validate failed : %v", tag, err)
				}
			} else {
				t.Logf("[SKIPPED: %s / out]", tag)
			}
		}

		if expect.err != nil {
			if result.out == nil {
				t.Fatalf("[ERROR  : %s / err] result is nil", tag)
			}
			if !bytes.Equal(expect.err, result.err) {
				t.Fatalf("[ERROR  : %s / err] Expect(%s) but returned(%s)", tag, expect.err, result.err)
			}

		} else {
			if expect.errFunc != nil {
				if err := expect.errFunc(result.err); err != nil {
					t.Fatalf("[ERROR  : %s / err] validate failed : %v", tag, err)
				}
			} else {
				t.Logf("[SKIPPED: %s / err]", tag)
			}
		}
	}

	if len(expect.execResults) != len(result.execResults) {
		t.Fatalf("[ERROR] len(expect.execResults)'%d' != len(result.execResults)'%d'",
			len(expect.execResults),
			len(result.execResults))
	}
	for execIndex, execExpected := range expect.execResults {
		execResult := result.execResults[execIndex]
		if execResult == nil {
			t.Fatalf("[ERROR] execResults[%d] is nil", execIndex)
		}

		assertUnit("compile", &execResult.compile, &execExpected.compile)
		assertUnit("compile", &execResult.link, &execExpected.link)

		// run
		if len(execExpected.run) != len(execResult.run) {
			t.Fatalf("[ERROR] len(execExpected.run)'%d' != len(execResult.run)'%d'",
				len(execExpected.run),
				len(execResult.run))
		}
		for execRunIndex, execRunExpected := range execExpected.run {
			execRunResult := execResult.run[execRunIndex]
			if execResult == nil {
				t.Fatalf("[ERROR] execResults[%d].run[%d] is nil", execIndex, execRunIndex)
			}

			assertUnit("compile", execRunResult, &execRunExpected)
		}
	}
}

func makeExecSpec(buildInst *BuildInstruction, runInsts []*RunInstruction) *ExecutionSpec {
	return &ExecutionSpec{
		BuildInst: buildInst,
		RunInsts:  runInsts,
	}
}

//
type testExpectedStatus struct {
	exited     BoolOptionalType
	exitStatus IntOptionalType
	signaled   BoolOptionalType
	signal     IntOptionalType
}
type testExpectedUnitResult struct {
	out, err         []byte
	outFunc, errFunc func(buf []byte) error
	status           *testExpectedStatus
}
type testExpectedExecResult struct {
	compile, link testExpectedUnitResult
	run           []testExpectedUnitResult
}
type testExpectedResult struct {
	execResults []testExpectedExecResult
}

//
type testUnitResult struct {
	out, err []byte
	result   *ExecutedResult
}
type testExecResult struct {
	compile, link testUnitResult
	run           map[int]*testUnitResult
}
type testResult struct {
	execResults map[int]*testExecResult
}

func makeHelperCallback(result *testResult) func(v interface{}) error {
	assumeExecResultHasValue := func(index int) {
		if result.execResults == nil {
			result.execResults = make(map[int]*testExecResult)
		}
		if result.execResults[index] == nil {
			result.execResults[index] = &testExecResult{}
		}
	}

	assumeUnitResultInRunHasValue := func(execResult *testExecResult, index int) {
		if execResult.run == nil {
			execResult.run = make(map[int]*testUnitResult)
		}
		if execResult.run[index] == nil {
			execResult.run[index] = &testUnitResult{}
		}
	}

	return func(v interface{}) error {
		switch v.(type) {
		case *StreamExecutedResult:
			r := v.(*StreamExecutedResult)

			assumeExecResultHasValue(r.MainIndex)
			execResult := result.execResults[r.MainIndex]

			switch r.Mode {
			case CompileMode:
				unitResult := &execResult.compile
				unitResult.result = r.Result
			case LinkMode:
				unitResult := &execResult.link
				unitResult.result = r.Result
			case RunMode:
				assumeUnitResultInRunHasValue(execResult, r.SubIndex)
				unitResult := execResult.run[r.SubIndex]

				unitResult.result = r.Result
			default:
				return errors.New("unsupported mode.")
			}

		case *StreamOutputResult:
			r := v.(*StreamOutputResult)

			assumeExecResultHasValue(r.MainIndex)
			execResult := result.execResults[r.MainIndex]

			switch r.Mode {
			case CompileMode:
				unitResult := &execResult.compile
				switch r.Output.Fd {
				case StdoutFd:
					unitResult.out = append(unitResult.out, r.Output.Buffer...)
				case StderrFd:
					unitResult.err = append(unitResult.err, r.Output.Buffer...)
				default:
					return errors.New("unsupported fd.")
				}

			case LinkMode:
				unitResult := &execResult.link
				switch r.Output.Fd {
				case StdoutFd:
					unitResult.out = append(unitResult.out, r.Output.Buffer...)
				case StderrFd:
					unitResult.err = append(unitResult.err, r.Output.Buffer...)
				default:
					return errors.New("unsupported fd.")
				}

			case RunMode:
				assumeUnitResultInRunHasValue(execResult, r.SubIndex)
				unitResult := execResult.run[r.SubIndex]

				switch r.Output.Fd {
				case StdoutFd:
					unitResult.out = append(unitResult.out, r.Output.Buffer...)
				case StderrFd:
					unitResult.err = append(unitResult.err, r.Output.Buffer...)
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
	Exists bool
}

type IntOptionalType struct {
	OptionalBase
	Value int
}

func IntOpt(v int) IntOptionalType {
	return IntOptionalType{
		OptionalBase: OptionalBase{true},
		Value:        v,
	}
}

type BoolOptionalType struct {
	OptionalBase
	Value bool
}

func BoolOpt(v bool) BoolOptionalType {
	return BoolOptionalType{
		OptionalBase: OptionalBase{true},
		Value:        v,
	}
}
