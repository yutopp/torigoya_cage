//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

// +build linux

package torigoya

import(
	"fmt"
	"errors"
	_ "strconv"
	"os"
	_ "os/user"
	_ "path/filepath"
    "os/exec"
)

type PackageUpdater interface {
	Update() error
}

type Context struct {
	basePath			string
	userFilesBasePath	string

	sandboxExecutor		SandboxExecutor

	procConfPath		string
	procConfTable		ProcConfigTable

	procSrcZipAddress	string
	packageUpdater		PackageUpdater
}

type MountOption struct {
	HostPath	string
	GuestPath	string
	IsReadOnly	bool
}

type CopyOption struct {
	HostPath	string
	GuestPath	string
}

type ResourceLimit struct {
	Core		uint64	// number
	Nofile		uint64	// number
	NProc		uint64	// number
	MemLock		uint64	// number
	CpuTime		uint64	// seconds
	Memory		uint64	// bytes
	FSize		uint64	// bytes
}

type SandboxExecutionOption struct {
	Mounts			[]MountOption
	Copies			[]CopyOption
	GuestHomePath	string
	Limits			*ResourceLimit
	Args			[]string
	Envs			[]string
}

type ExecuteCallBackType	func(*StreamOutput)
type SandboxExecutor interface {
	Execute(*SandboxExecutionOption, *os.File, ExecuteCallBackType) (*ExecutedResult, error)
}




type AwahoSandboxExecutor struct {
	ExecutablePath		string
}




type ContextOptions struct {
	BasePath				string
	UserFilesBasePath		string

	SandboxExec				SandboxExecutor

	ProcConfigPath			string
	ProcSrcZipAddress		string
	PackageUpdater			PackageUpdater
}

func InitContext(opts *ContextOptions) (*Context, error) {
	// TODO: change to checking capability
	expectRoot()

	// create ~~~ Directory, if not existed
	if !fileExists(opts.UserFilesBasePath) {
		err := os.Mkdir(opts.UserFilesBasePath, os.ModeDir | 0700)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Couldn't create directory %s", opts.UserFilesBasePath))
		}
	}

	// LoadProcConfigTable
	proc_conf_table, err := LoadProcConfigs(opts.ProcConfigPath)
	if err != nil {
		// make no error if table coulnd't be loaded
		proc_conf_table = nil
	}

	//
	return &Context{
		basePath:			opts.BasePath,
		userFilesBasePath:	opts.UserFilesBasePath,

		sandboxExecutor:	opts.SandboxExec,

		procConfPath:		opts.ProcConfigPath,
		procConfTable:		proc_conf_table,
		procSrcZipAddress:	opts.ProcSrcZipAddress,
		packageUpdater:		opts.PackageUpdater,
	}, nil
}


func (ctx *Context) HasProcTable() bool {
	return ctx.procConfTable != nil
}


func (ctx *Context) UpdatePackages() error {
	if ctx.packageUpdater == nil {
		return errors.New("Package Updater was not registerd")
	}

	err := ctx.packageUpdater.Update()

	// TODO: fix it
    fmt.Printf("= /usr/local/torigoya ============================\n")
	out, err := exec.Command("/bin/ls", "-la", "/usr/local/torigoya").Output()
	if err != nil {
		fmt.Printf("error:: %s\n", err.Error())
	} else {
		fmt.Printf("package update passed:: %s\n", out)
	}
	fmt.Printf("==================================================\n")

    return err
}


func (ctx *Context) ReloadProcTable() error {
	// RELOAD LoadProcConfigTable
	proc_conf_table, err := LoadProcConfigs(ctx.procConfPath)
	if err != nil {
		return err
	}

	// rewrite
	ctx.procConfTable = proc_conf_table

	return nil
}

func (ctx *Context) UpdateProcTable() error {
	if ctx.procSrcZipAddress != "" {
		if err := ctx.procConfTable.UpdateFromWeb(ctx.procSrcZipAddress, ctx.basePath); err != nil {
			return err
		}
	}

	return ctx.ReloadProcTable()
}
