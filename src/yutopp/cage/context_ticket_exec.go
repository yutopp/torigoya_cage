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
	"log"
	"fmt"
	"errors"
	"os"
	"sync"
	"io/ioutil"

	"github.com/jmcvetta/randutil"
)

const (
	CompileMode = iota
	LinkMode
	RunMode
)

const guestHome = "/home/torigoya"

var (
	compileFailedError	= errors.New("compile failed")
	linkFailedError		= errors.New("link failed")
	buildFailedError	= errors.New("build failed")
)

//
type invokeResultRecieverCallback		func(interface{}) error


// ========================================
// ========================================
func (ctx *Context) ExecTicket(
	ticket				*Ticket,
	callback			invokeResultRecieverCallback,
) error {
	base_name, err := func() (string, error) {
		if ticket.BaseName != "" {
			return ticket.BaseName, nil
		} else {
			return randutil.AlphaString(32)
		}
	}()
	if err != nil {
		return err
	}

	log.Printf("-- START ticket => %s\n", base_name)
	defer log.Printf("-- FINISH ticket  => %s\n", base_name)

	log.Printf("mapSources => %s\n", base_name)
	// map files and create user's directory
	path_used_as_home, err := ctx.mapSources(base_name, ticket.Sources)
	if err != nil {
		return err
	}
	log.Printf("basePath => %s\n", path_used_as_home)

	// build
	if ticket.IsBuildRequired() {
		if err := ctx.execBuild(path_used_as_home, ticket.BuildInst, callback); err != nil {
			if err == buildFailedError {
				return nil	// not an error
			} else {
				return err
			}
		}
	}

	// run
	if errs := ctx.execManagedRun(path_used_as_home, ticket.RunInst, callback); errs != nil {
		// TODO: process error
		var s string
		for err := range errs {
			log.Printf("exec error %v\n", err)
			s += fmt.Sprintf("%v: ", err)
		}
		return errors.New("Failed to exec inputs : " + s)
	}

	return nil
}


// ========================================
// ========================================
func (ctx *Context) execBuild(
	path_used_as_home	string,
	build_inst			*BuildInstruction,
	callback			invokeResultRecieverCallback,
) error {
	log.Printf("$$$$$$$$$$ START build => %s\n", path_used_as_home)
	defer log.Printf("$$$$$$$$$$ FINISH build  => %s\n", path_used_as_home)

	// compile phase
	log.Printf("$$$$$$$$$$ START compile\n")
	if err := ctx.invokeCompileCommand(
		path_used_as_home,
		build_inst.CompileSetting,
		callback,
	); err != nil {
		if err == compileFailedError {
			return buildFailedError
		} else {
			return err
		}
	}

	// link phase :: if link command is separated, so call linking commands
	if build_inst.IsLinkIndependent() {
		if err := ctx.invokeLinkCommand(
			path_used_as_home,
			build_inst.LinkSetting,
			callback,
		); err != nil {
			if err == linkFailedError {
				return buildFailedError
			} else {
				return err
			}
		}
	}

	return nil
}

func (ctx *Context) execManagedRun(
	path_used_as_home	string,
	run_inst			*RunInstruction,
	callback			invokeResultRecieverCallback,
) []error {
	log.Printf("$$$$$$$$$$ START run => %s\n", path_used_as_home)
	defer log.Printf("$$$$$$$$$$ FINISH run => %s\n", path_used_as_home)

	//
	wg := new(sync.WaitGroup)
	m := new(sync.Mutex)
	var errs []error = nil

	// execute inputs async
	for _index, _input := range run_inst.Inputs {
		wg.Add(1)

		go func(index int, input Input) {
			defer wg.Done()

			if err := ctx.invokeRunCommand(
				path_used_as_home,
				index,
				&input,
				callback,
			); err != nil {
				m.Lock()
				if errs == nil { errs = make([]error, 0) }
				errs = append(errs, err)
				m.Unlock()
			}
		}(_index, _input)
	}
	wg.Wait()

	return errs
}


// ========================================
// ========================================
func (ctx *Context) invokeCompileCommand(
	path_used_as_home	string,
	exec_inst			*ExecutionSetting,
	callback			invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeCompileCommand")

	opts := &SandboxExecutionOption{
		Mounts:	[]MountOption{
			MountOption{
				HostPath: path_used_as_home,
				GuestPath: guestHome,
				IsReadOnly: false,
				DoChown: true,
			},
			MountOption{
				HostPath: ctx.packageInstalledBasePath,
				GuestPath: ctx.packageInstalledBasePath,
				IsReadOnly: true,
				DoChown: false,
			},
		},
		GuestHomePath: guestHome,
		Limits: &ResourceLimit{
			Core: 0,							// Process can NOT create CORE file
			Nofile: 512,						// Process can open 512 files
			NProc: 30,							// Process can create processes to 30
			MemLock: 1024,						// Process can lock 1024 Bytes by mlock(2)
			CpuTime: exec_inst.CpuTimeLimit,	// sec
			Memory: exec_inst.MemoryBytesLimit,	// bytes
			FSize: 5 * 1024 * 1024,				// Process can writes a file only 5MiB
		},
		Args: exec_inst.Args,
		Envs: exec_inst.Envs,
	}

	f := func(output *StreamOutput) error {
		return callback(&StreamOutputResult{
			Mode: CompileMode,
			Index: 0,
			Output: output,
		})
	}
	result, err := ctx.sandboxExecutor.Execute(opts, nil, f)
	log.Printf("Compile Executed >> %v / %v", result, err)
	if err != nil {
		return err
	}

	callback(&StreamExecutedResult{
		Mode: CompileMode,
		Index: 0,
		Result: result,
	})

	if !result.IsSucceeded() {
		return compileFailedError
	} else {
		return nil
	}
}

func (ctx *Context) invokeLinkCommand(
	path_used_as_home	string,
	exec_inst			*ExecutionSetting,
	callback			invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeLinkCommand")

	opts := &SandboxExecutionOption{
		Mounts:	[]MountOption{
			MountOption{
				HostPath: path_used_as_home,
				GuestPath: guestHome,
				IsReadOnly: false,
				DoChown: true,
			},
			MountOption{
				HostPath: ctx.packageInstalledBasePath,
				GuestPath: ctx.packageInstalledBasePath,
				IsReadOnly: true,
				DoChown: false,
			},
		},
		GuestHomePath: guestHome,
		Limits: &ResourceLimit{
			Core: 0,							// Process can NOT create CORE file
			Nofile: 512,						// Process can open 512 files
			NProc: 30,							// Process can create processes to 30
			MemLock: 1024,						// Process can lock 1024 Bytes by mlock(2)
			CpuTime: 10,						// 10 sec
			Memory: 2 * 1024 * 1024 * 1024,		// 2GiB[fixed]
			FSize: 40 * 1024 * 1024,			// 40MiB[fixed]
		},
		Args: exec_inst.Args,
		Envs: exec_inst.Envs,
	}

	f := func(output *StreamOutput) error {
		return callback(&StreamOutputResult{
			Mode: LinkMode,
			Index: 0,
			Output: output,
		})
	}
	result, err := ctx.sandboxExecutor.Execute(opts, nil, f)
	log.Printf("Link Executed >> %v / %v", result, err)
	if err != nil {
		return err
	}

	callback(&StreamExecutedResult{
		Mode: LinkMode,
		Index: 0,
		Result: result,
	})

	if !result.IsSucceeded() {
		return linkFailedError
	} else {
		return nil
	}
}

func (ctx *Context) invokeRunCommand(
	path_used_as_home	string,
	index				int,
	input				*Input,
	callback			invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeRunInputCommand")

	stdin := input.Stdin
	exec_inst := input.RunSetting

	var temp_stdin *os.File = nil
	if stdin != nil {
		log.Println("use stdin")

		f, err := ioutil.TempFile("", "torigoya-inputs-");
		if err != nil { return err }
		defer f.Close()

		n, err := f.Write(stdin.Data)
		if err != nil {
			return err
		}
		if n != len(stdin.Data) {
			return errors.New("input length is different")
		}

		noff, err := f.Seek(0, os.SEEK_SET)
		if err != nil {
			return err
		}
		if noff != 0 {
			return errors.New("offset is not 0")
		}

		// r--/---/---
		if err := f.Chmod(0400); err != nil {
			return err
		}

		temp_stdin = f
	}

	opts := &SandboxExecutionOption{
		Mounts: []MountOption{
			MountOption{
				HostPath: ctx.packageInstalledBasePath,
				GuestPath: ctx.packageInstalledBasePath,
				IsReadOnly: true,
				DoChown: false,
			},
		},
		Copies:	[]CopyOption{	// NOTE: NOT "Mount", to run async
			CopyOption{
				HostPath: path_used_as_home,
				GuestPath: guestHome,
			},
		},
		GuestHomePath: guestHome,
		Limits: &ResourceLimit{
			Core: 0,							// Process can NOT create CORE file
			Nofile: 512,						// Process can open 512 files
			NProc: 30,							// Process can create processes to 30
			MemLock: 1024,						// Process can lock 1024 Bytes by mlock(2)
			CpuTime: exec_inst.CpuTimeLimit,	// sec
			Memory: exec_inst.MemoryBytesLimit,	// bytes
			FSize: 1 * 1024 * 1024,				// Process can writes a file only 1MB
		},
		Args: exec_inst.Args,
		Envs: exec_inst.Envs,
	}

	f := func(output *StreamOutput) error {
		return callback(&StreamOutputResult{
			Mode: RunMode,
			Index: index,
			Output: output,
		})
	}
	result, err := ctx.sandboxExecutor.Execute(opts, temp_stdin, f)
	log.Printf("Run Executed >> %v / %v", result, err)
	if err != nil {
		return err
	}

	log.Printf(">> %v", result)
	callback(&StreamExecutedResult{
		Mode: RunMode,
		Index: index,
		Result: result,
	})
	log.Println("sandboxExecutor.Exit >> %v", result)

	return nil
}


// ========================================
// ========================================
func (ctx *Context) mapSources(
	base_name			string,
	sources				[]*SourceData,
) (string, error) {
	// unpack source codes
	source_contents, err := convertSourcesToContents(sources)
	if err != nil {
		return "", err
	}

	//
	user_home_dir_path, err := ctx.createMultipleTargets(
		base_name,
		source_contents,
	)
	if err != nil {
		return "", errors.New("couldn't create multi target : " + err.Error());
	}

	return user_home_dir_path, nil
}
