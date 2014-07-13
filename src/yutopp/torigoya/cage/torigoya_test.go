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



func TestTicket(t *testing.T) {
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
	f := makehelperCallback(&result)
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
			out: "hello!\ninput is 0\n",
		},
		1: &test_result_unit{
			out: "hello!\ninput is 100\n",
		},
		},
	}

	//
	assertTestResult(t, &result, &expect_result)
}

func assertTestResult(t *testing.T, result, expect *test_result) {
	assertUnit := func (tag string, result, expect *test_result_unit) {
		if expect.out != result.out {
			t.Fatalf("[%s / out] Expect(%s) but returned(%s)", tag, expect.out, result.out)
		}

		if expect.err != result.err {
			t.Fatalf("[%s / err] Expect(%s) but returned(%s)", tag, expect.err, result.err)
		}
	}

	assertUnit("compile", &result.compile, &expect.compile)
	assertUnit("link", &result.link, &expect.link)

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
	out, err	string
	result		*ExecutedResult
}
type test_result struct {
	compile, link	test_result_unit
	run				map[int]*test_result_unit
}

func makehelperCallback(result *test_result) func(v interface{}) {
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
					result.compile.out = result.compile.out + string(r.Output.Buffer)
				case StderrFd:
					result.compile.err = result.compile.err + string(r.Output.Buffer)
				}

			case LinkMode:
				switch r.Output.Fd {
				case StdoutFd:
					result.link.out = result.link.out + string(r.Output.Buffer)
				case StderrFd:
					result.link.err = result.link.err + string(r.Output.Buffer)
				}

			case RunMode:
				if result.run[r.Index] == nil { result.run[r.Index] = &test_result_unit{} }
				switch r.Output.Fd {
				case StdoutFd:
					result.run[r.Index].out = result.run[r.Index].out + string(r.Output.Buffer)
				case StderrFd:
					result.run[r.Index].err = result.run[r.Index].err + string(r.Output.Buffer)
				}
			}

		default:
			panic("unsupported type.");
		}
	}
}
