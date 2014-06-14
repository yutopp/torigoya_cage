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
	source_name := "prog.cpp"
	content := "test test test"
	group_id := 1000

	source_full_path, err := ctx.createTarget(base_name, group_id, source_name, func(f *os.File) (error) {
		// write content
		_, err := f.WriteString(content)
		return err
	})
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	//
	t.Logf("source path: %s", source_full_path)
	if source_full_path != filepath.Join(ctx.sandboxDir, base_name, ctx.jailedUserDir, source_name) {
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
	source_name := "prog2.cpp"
	content := "test test test2"
	group_id := 1000

	ctx.createTarget(base_name, group_id, source_name, func(f *os.File) (error) {
		_, err := f.WriteString(content)
		return err
	})

	user_dir_path, _, err := ctx.reassignTarget(base_name, group_id, func(s string) (string, error) {return "", nil})
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
	source_name := "prog2.cpp"
	content := "test test test2"
	group_id := 1000

	ctx.createTarget(base_name, group_id, source_name, func(f *os.File) (error) {
		_, err := f.WriteString(content)
		return err
	})

	user_dir_path, _, err := ctx.reassignTarget(base_name, group_id, func(s string) (string, error) {return "", nil})
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


	_, _, err = ctx.reassignTarget(base_name, group_id, func(s string) (string, error) {return "", nil})
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
	source_name := "prog2.cpp"
	content := "test test test2"
	group_id := 1000

	stdin_name := "in" + strconv.FormatInt(time.Now().Unix(), 10)
	stdin_content := "iniini~~~"

	ctx.createTarget(base_name, group_id, source_name, func(f *os.File) (error) {
		_, err := f.WriteString(content)
		return err
	})

	user_dir_path, input_path, err := ctx.reassignTarget(base_name, group_id, func(base_directory_name string) (string, error) {
		return ctx.createInput(base_directory_name, group_id, stdin_name, stdin_content)
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

	//
	t.Logf("input path: %s", input_path)
	if input_path != filepath.Join(ctx.sandboxDir, base_name, ctx.jailedUserDir, "stdin", stdin_name) {
		t.Errorf(fmt.Sprintf("%s", input_path))
		return
	}

}


func TestinvokeProcessClonerBase(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	err := invokeProcessClonerBase(gopath, "process_cloner", nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
}


func TestBootStrap(t *testing.T) {
	err := sandboxBootstrap(nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
}

func TestBuild(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath)
	if err != nil {
		t.Errorf(err.Error())
		return
	}


	base_name := "aaa4" + strconv.FormatInt(time.Now().Unix(), 10)
	sources := []SourceData{
		SourceData{
		},
	}

	// build
	if err := ctx.build(base_name, sources); err != nil {
		t.Errorf(err.Error())
		return
	}

}
