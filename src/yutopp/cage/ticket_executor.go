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
	"io/ioutil"
	_ "path/filepath"
)


// send this message to sandbox process
type ExecMessage struct {
	Profile				*ProcProfile
	StdinFilePath		*string
	Setting				*ExecutionSetting
	Mode				int
}

const (
	CompileMode = iota
	LinkMode
	RunMode
)

var (
	compileFailedError	= errors.New("compile failed")
	linkFailedError		= errors.New("link failed")
	buildFailedError	= errors.New("build failed")
)

// ========================================
func (ctx *Context) ExecTicket(
	ticket				*Ticket,
	callback			invokeResultRecieverCallback,
) error {
	log.Printf("-- START ticket => %s\n", ticket.BaseName)
	defer log.Printf("-- FINISH ticket  => %s\n", ticket.BaseName)

	log.Printf("mapSources => %s\n", ticket.BaseName)
	// map files and create user's directory
	path_used_as_home, err := ctx.mapSources(ticket.BaseName, ticket.Sources)
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


//
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
	var errs []error = nil
	// ========================================
	for index, input := range run_inst.Inputs {
		// TODO: async
		err := ctx.invokeRunCommand(
			path_used_as_home,
			index,
			&input,
			callback,
		)

		if err != nil {
			if errs == nil { errs = make([]error, 0) }
			errs = append(errs, err)
		}
	}

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

	const GuestHome = "/home/torigoya"

	opts := &SandboxExecutionOption{
		Mounts:	[]MountOption{
			MountOption{
				HostPath: path_used_as_home,
				GuestPath: GuestHome,
				IsReadOnly: false,
			},
		},
		GuestHomePath: GuestHome,
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

	f := func(output *StreamOutput) {
		callback(&StreamOutputResult{
			Mode: CompileMode,
			Index: 0,
			Output: output,
		})
	}
	result, err := ctx.sandboxExecutor.Execute(opts, nil, f)
	log.Println(">> %v", result)
	if err != nil {
		return err
	}

	callback(&StreamExecutedResult{
		Mode: CompileMode,
		Index: 0,
		Result: result,
	})

	return err
}


func (ctx *Context) invokeLinkCommand(
	path_used_as_home	string,
	exec_inst			*ExecutionSetting,
	callback			invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeLinkCommand")

	const GuestHome = "/home/torigoya"

	opts := &SandboxExecutionOption{
		Mounts:	[]MountOption{
			MountOption{
				HostPath: path_used_as_home,
				GuestPath: GuestHome,
				IsReadOnly: false,
			},
		},
		GuestHomePath: GuestHome,
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

	f := func(output *StreamOutput) {
		callback(&StreamOutputResult{
			Mode: LinkMode,
			Index: 0,
			Output: output,
		})
	}
	result, err := ctx.sandboxExecutor.Execute(opts, nil, f)
	log.Println(">> %v", result)
	if err != nil {
		return err
	}

	callback(&StreamExecutedResult{
		Mode: LinkMode,
		Index: 0,
		Result: result,
	})

	return err
}

func (ctx *Context) invokeRunCommand(
	path_used_as_home	string,
	index				int,
	input				*Input,
	callback			invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeRunInputCommand")

	const GuestHome = "/home/torigoya"
	const GuestInputs = "/home/torigoya/inputs"

	// stdin := input.stdin
	exec_inst := input.setting

	var temp_stdin *os.File = nil
	if input.stdin != nil {
		log.Println("use stdin")

		f, err := ioutil.TempFile("", "torigoya-inputs-");
		if err != nil { return err }
		defer f.Close()

		n, err := f.Write(input.stdin.Data)
		if err != nil {
			return err
		}
		if n != len(input.stdin.Data) {
			return errors.New("input length is different")
		}

		noff, err := f.Seek(0, os.SEEK_SET)
		if err != nil {
			return err
		}
		if noff != 0 {
			return errors.New("offseet is not 0")
		}

		// r--/---/---
		if err := f.Chmod(0400); err != nil {
			return err
		}

		temp_stdin = f
	}

	opts := &SandboxExecutionOption{
		Mounts:	[]MountOption{
			MountOption{
				HostPath: path_used_as_home,
				GuestPath: GuestHome,
				IsReadOnly: false,
			},
		},
		GuestHomePath: GuestHome,
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

	f := func(output *StreamOutput) {
		callback(&StreamOutputResult{
			Mode: RunMode,
			Index: index,
			Output: output,
		})
	}
	result, err := ctx.sandboxExecutor.Execute(opts, temp_stdin, f)
	log.Println(">> %v", result)
	if err != nil {
		return err
	}

	callback(&StreamExecutedResult{
		Mode: RunMode,
		Index: index,
		Result: result,
	})

	return err
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


// ========================================
// ========================================


//
type StreamOutputResult struct {
	Mode		int
	Index		int
	Output		*StreamOutput
}

func (r *StreamOutputResult) ToTuple() []interface{} {
	return []interface{}{ r.Mode, r.Index, r.Output.ToTuple() }
}


//
type StreamExecutedResult struct {
	Mode		int
	Index		int
	Result		*ExecutedResult
}
func (r *StreamExecutedResult) ToTuple() []interface{} {
	return []interface{}{ r.Mode, r.Index, r.Result.ToTuple() }
}


//
type invokeResultRecieverCallback		func(interface{})
