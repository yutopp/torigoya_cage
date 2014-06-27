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
	ctx, err := InitContext(gopath)
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


func TestReassignTarget(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	base_name := "aaa2" + strconv.FormatInt(time.Now().Unix(), 10)
	content := &TextContent{
		"prog.cpp",
		[]byte("test test test"),
	}
	group_id := 1000

	ctx.createTarget(base_name, group_id, content)

	user_dir_path, _, err := ctx.reassignTarget(
		base_name,
		group_id,
		func(s string) ([]string, error) { return nil, nil },
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
	ctx, err := InitContext(gopath)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	base_name := "aaa2" + strconv.FormatInt(time.Now().Unix(), 10)
	content := &TextContent{
		"prog.cpp",
		[]byte("test test test"),
	}
	group_id := 1000

	ctx.createTarget(base_name, group_id, content)

	user_dir_path, _, err := ctx.reassignTarget(
		base_name,
		group_id,
		func(s string) ([]string, error) { return nil, nil },
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


	_, _, err = ctx.reassignTarget(base_name, group_id, func(s string) ([]string, error) { return nil, nil })
	if err != nil {
		t.Errorf(err.Error())
		return
	}
}


func TestCreateInput(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	base_name := "aaa3" + strconv.FormatInt(time.Now().Unix(), 10)
	content := &TextContent{
		"prog.cpp",
		[]byte("test test test"),
	}
	group_id := 1000

	stdin := &TextContent{
		"in" + strconv.FormatInt(time.Now().Unix(), 10),
		[]byte("iniini~~~"),
	}

	ctx.createTarget(base_name, group_id, content)

	user_dir_path, input_paths, err := ctx.reassignTarget(base_name, group_id, func(base_directory_name string) ([]string, error) {
		path, err := ctx.createInput(base_directory_name, group_id, stdin)
		if err != nil { return nil, err }
		return []string{ path }, nil
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

	if len(input_paths) != 1 {
		t.Errorf("length of input_paths should be 1 (%d)", len(input_paths))
		return
	}

	//
	t.Logf("input path: %s", input_paths[0])
	if input_paths[0] != filepath.Join(ctx.sandboxDir, base_name, ctx.jailedUserDir, "stdin", stdin.Name) {
		t.Errorf("%v", input_paths)
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




func TestBuild(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath)
	if err != nil {
		t.Errorf(err.Error())
		return
	}


	base_name := "aaa4" + strconv.FormatInt(time.Now().Unix(), 10)
	sources := []*SourceData{
		&SourceData{
			"test.cpp",
			[]byte(""),
			false,
		},
	}

	proc_profile := &ProcProfile{
		IsBuildRequired: true,
		IsLinkIndependent: true,
	}

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

	f := func(v interface{}) {
		t.Logf("%V", v)
	}

	// build
	if err := ctx.invokeBuild(base_name, sources, proc_profile, build_inst, f); err != nil {
		t.Errorf(err.Error())
		return
	}
}


func TestAAA(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath)
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
			[]byte(""),
			false,
		},
	}

	// load id:0/version:0.0.0
	configs, _ := LoadProcConfigs(filepath.Join(gopath, "test_proc_profiles"))
	proc_profile := configs[0].Versioned["0.0.0"]

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
	if err := ctx.invokeBuild(base_name, sources, &proc_profile, build_inst, nil); err != nil {
		t.Errorf(err.Error())
		return
	}
}
