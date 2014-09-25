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
	"path/filepath"
	"time"
	"strconv"
	"bytes"

	"sync"
)


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
	group_id := 1000

	source_full_path, err := ctx.createTarget(base_name, group_id, content)
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
	group_id := 1000

	for i:=0; i<2; i++ {
		if _, err := ctx.createTarget(base_name, group_id, content); err != nil {
			t.Fatalf(err.Error())
		}
	}
}




func TestReassignTarget(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	base_name := "aaa2" + strconv.FormatInt(time.Now().Unix(), 10)
	content := &TextContent{
		"prog.cpp",
		[]byte("test test test"),
	}
	user_id := 1000
	group_id := 1000

	ctx.createTarget(base_name, group_id, content)

	user_dir_path, _, err := ctx.reassignTarget(
		base_name,
		user_id,
		group_id,
		func(s string) (*string, error) { return nil, nil },
	)
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
}


func TestReassignTarget2(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	base_name := "aaa2" + strconv.FormatInt(time.Now().Unix(), 10)
	content := &TextContent{
		"prog.cpp",
		[]byte("test test test"),
	}
	user_id := 1000
	group_id := 1000

	ctx.createTarget(base_name, group_id, content)

	user_dir_path, _, err := ctx.reassignTarget(
		base_name,
		user_id,
		group_id,
		func(s string) (*string, error) { return nil, nil },
	)
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


	_, _, err = ctx.reassignTarget(base_name, user_id, group_id, func(s string) (*string, error) { return nil, nil })
	if err != nil {
		t.Errorf(err.Error())
		return
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

	ctx.createTarget(base_name, group_id, content)

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


/*
func TestInvokeProcessClonerBase(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	err := invokeProcessClonerBase(filepath.Join(gopath, "bin"), "process_cloner", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
}
*/


func TestBootStrap(t *testing.T) {
	err := runAsManagedUser(nil)
	if err != nil {
		t.Errorf("TestBootStrap" + err.Error())
		return
	}
}

/*
func TestExec(t *testing.T) {
	limit := &ResourceLimit{
		CPU: 10,		// CPU can be used only cpu_limit_time(sec)
		AS: 1 * 1024 * 1024 * 1024,		// Memory can be used only memory_limit_bytes
		FSize: 5 * 1024 * 1024,				// Process can writes a file only 5 MBytes
	}


	err := execc(limit, "ls", []string{}, map[string]string{"PATH": "/bin"})
	if err != nil {
		t.Fatalf(err.Error())
	}

	t.Fatalf("ababa")
}

*/




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



func TestTicketBasic(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
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
	expect_result := test_result{
		compile: test_result_unit{
		},
		link: test_result_unit{
		},
		run: map[int]*test_result_unit{
			0: &test_result_unit{
				out: []byte("hello!\ninput is 0\n"),
			},
			1: &test_result_unit{
				out: []byte("hello!\ninput is 100\n"),
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}


func TestTicketBasicParallel1(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	var wg sync.WaitGroup
	const num = 5
	var fx [num]bool
	for i := 0; i < num; i++ {
		wg.Add(1)

		go func(no int) {
			defer func() {
				fmt.Printf("Done! %d\n", no)
				fx[no] = true
				fmt.Printf("fs! %v\n", fx)
				wg.Done()
			}()

			//
			base_name := "paralell1_no_" + strconv.Itoa(no) + "_" + strconv.FormatInt(time.Now().Unix(), 10)

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

			// execute
			var result test_result
			result.run = make(map[int]*test_result_unit)
			f := makeHelperCallback(&result)
			if err := ctx.ExecTicket(ticket, f); err != nil {
				t.Errorf(err.Error())
				return
			}

			//
			expect_result := test_result{
				compile: test_result_unit{
				},
				link: test_result_unit{
				},
				run: map[int]*test_result_unit{
					0: &test_result_unit{
						out: []byte("hello!\ninput is 0\n"),
					},
					1: &test_result_unit{
						out: []byte("hello!\ninput is 100\n"),
					},
				},
			}

			//
			t.Logf("%d =====\n", no)
			t.Logf("%V\n", result)
			assertTestResult(t, &result, &expect_result)
		}(i)
	}

	wg.Wait()
}


func TestTicketBasicParallel2(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	var wg sync.WaitGroup
	const num = 30
	var fx [num]bool
	for i := 0; i < num; i++ {
		wg.Add(1)

		go func(no int) {
			defer func() {
				fmt.Printf("Done! %d\n", no)
				fx[no] = true
				fmt.Printf("fs! %v\n", fx)
				wg.Done()
			}()

			//
			base_name := "paralell2_no_" + strconv.Itoa(no) + "_" + strconv.FormatInt(time.Now().Unix(), 10)

			//
			sources := []*SourceData{
				&SourceData{
					"prog.cpp",
					[]byte(`
#include <iostream>
#include <unistd.h>

int main() {
	std::cout << "hello!" << std::endl;
	int i;
	std::cin >> i;
	usleep(200000); // 0.2 sec
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

			// execute
			var result test_result
			result.run = make(map[int]*test_result_unit)
			f := makeHelperCallback(&result)
			if err := ctx.ExecTicket(ticket, f); err != nil {
				t.Errorf(err.Error())
				return
			}



			//
			expect_result := test_result{
				compile: test_result_unit{
				},
				link: test_result_unit{
				},
				run: map[int]*test_result_unit{
					0: &test_result_unit{
						out: []byte("hello!\ninput is 0\n"),
					},
					1: &test_result_unit{
						out: []byte("hello!\ninput is 100\n"),
					},
				},
			}

			//
			t.Logf("%d =====\n", no)
			t.Logf("%V\n", result)
			assertTestResult(t, &result, &expect_result)
		}(i)
	}

	wg.Wait()
}



func TestTicketTLE(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	base_name := "aaa7" + strconv.FormatInt(time.Now().Unix(), 10)

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
			CpuTimeLimit: 3,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			CpuTimeLimit: 3,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				stdin: nil,
				setting: &ExecutionSetting{
					CpuTimeLimit: 1,
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
	expect_result := test_result{
		compile: test_result_unit{
		},
		link: test_result_unit{
		},
		run: map[int]*test_result_unit{
			0: &test_result_unit{
				result: &ExecutedResult {
					Status: CPULimit,
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}


func TestTicketTLEWithHandling(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	base_name := "tle_with_handling" + strconv.FormatInt(time.Now().Unix(), 10)

	//
	sources := []*SourceData{
		&SourceData{
			"prog.cpp",
			[]byte(`
#include <iostream>
#include <signal.h>

void foo(int _unused)
{}

int main() {
	signal(SIGXCPU, foo);
	for(;;);
}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			CpuTimeLimit: 3,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			CpuTimeLimit: 3,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				stdin: nil,
				setting: &ExecutionSetting{
					CpuTimeLimit: 1,
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
	expect_result := test_result{
		compile: test_result_unit{
		},
		link: test_result_unit{
		},
		run: map[int]*test_result_unit{
			0: &test_result_unit{
				result: &ExecutedResult {
					Status: CPULimit,
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}


func TestTicketTLEWithSleep(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	base_name := "tle_with_sleep" + strconv.FormatInt(time.Now().Unix(), 10)

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
			CpuTimeLimit: 3,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			CpuTimeLimit: 3,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				stdin: nil,
				setting: &ExecutionSetting{
					CpuTimeLimit: 1,
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
	expect_result := test_result{
		compile: test_result_unit{
		},
		link: test_result_unit{
		},
		run: map[int]*test_result_unit{
			0: &test_result_unit{
				result: &ExecutedResult {
					Status: Error,
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}



func TestTicketMLE(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	base_name := "mle" + strconv.FormatInt(time.Now().Unix(), 10)

	//
	sources := []*SourceData{
		&SourceData{
			"prog.cpp",
			[]byte(`
#include <iostream>

int main() {

}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			CpuTimeLimit: 3,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			CpuTimeLimit: 3,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
	}

	//
	run_inst := &RunInstruction{
		Inputs: []Input{
			Input{
				stdin: nil,
				setting: &ExecutionSetting{
					CpuTimeLimit: 1,
					MemoryBytesLimit: 3 * 1024 * 1024,
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
	expect_result := test_result{
		compile: test_result_unit{
		},
		link: test_result_unit{
		},
		run: map[int]*test_result_unit{
			0: &test_result_unit{
				result: &ExecutedResult {
					Status: CPULimit,
				},
			},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}



func TestTicketRepeat(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath, "root", filepath.Join(gopath, "files", "proc_profiles_for_core_test"), "", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	base_name := "aaa8" + strconv.FormatInt(time.Now().Unix(), 10)

	//
	sources := []*SourceData{
		&SourceData{
			"prog.cpp",
			[]byte(`
#include <iostream>

int main() {
	for(int i=0; i<200000; ++i) std::cout << i << "\n" << std::flush;
}
`),
			false,
		},
	}

	//
	build_inst := &BuildInstruction{
		CompileSetting: &ExecutionSetting{
			CpuTimeLimit: 3,
			MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
		},
		LinkSetting: &ExecutionSetting{
			CpuTimeLimit: 3,
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

	// execute
	var result test_result
	result.run = make(map[int]*test_result_unit)
	f := makeHelperCallback(&result)
	if err := ctx.ExecTicket(ticket, f); err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	out := []byte{}
	for i:=0; i<200000; i++ {
		out = append(out, fmt.Sprintf("%d\n", i)...)
	}

	expect_result := test_result{
		compile: test_result_unit{
		},
		link: test_result_unit{
		},
		run: map[int]*test_result_unit{
			0: &test_result_unit{
				out: out,
				result: &ExecutedResult {
					Status: Passed,
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
func assertTestResult(t *testing.T, result, expect *test_result) {
	assertUnit := func (tag string, result, expect *test_result_unit) {
		if expect.result != nil {
			if result.result == nil {
				t.Fatalf("[ERROR  : %s / result] result is nil", tag)
			}
			if expect.result.Status != result.result.Status {
				t.Fatalf("ERROR  : [%s / result.Status] Expect(%s) but returned(%s)", tag,
					expect.result.Status,
					result.result.Status,
				)
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

type test_result_unit struct {
	out, err	[]byte
	result		*ExecutedResult
}
type test_result struct {
	compile, link	test_result_unit
	run				map[int]*test_result_unit
}

func makeHelperCallback(result *test_result) func(v interface{}) {
	return func(v interface{}) {
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
				}

			case LinkMode:
				switch r.Output.Fd {
				case StdoutFd:
					result.link.out = append(result.link.out, r.Output.Buffer...)
				case StderrFd:
					result.link.err = append(result.link.err, r.Output.Buffer...)
				}

			case RunMode:
				if result.run[r.Index] == nil { result.run[r.Index] = &test_result_unit{} }
				switch r.Output.Fd {
				case StdoutFd:
					result.run[r.Index].out = append(result.run[r.Index].out, r.Output.Buffer...)
				case StderrFd:
					result.run[r.Index].err = append(result.run[r.Index].err, r.Output.Buffer...)
				}
			}

		default:
			panic("unsupported type.");
		}
	}
}
