//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

// +build linux

package torigoya

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

const (
	CompileMode = iota
	LinkMode
	RunMode
)

const guestHome = "/home/torigoya"

var (
	compileFailedError = errors.New("compile failed")
	linkFailedError    = errors.New("link failed")
	buildFailedError   = errors.New("build failed")
)

//
type invokeResultRecieverCallback func(interface{}) error

// ========================================
// ========================================
func (ctx *Context) ExecTicket(
	ticket *Ticket,
	callback invokeResultRecieverCallback,
) error {
	baseName := ticket.BaseName
	if baseName == "" {
		return errors.New("ticket.BaseName cannot be empty")
	}

	log.Printf("-- START ticket => %s\n", baseName)
	defer log.Printf("-- FINISH ticket  => %s\n", baseName)

	m := new(sync.Mutex)
	errs := map[int]error{}

	// exec
	wg := new(sync.WaitGroup)
	for index_, execSpec_ := range ticket.ExecSpecs {
		wg.Add(1)
		go func(index int, execSpec *ExecutionSpec) {
			defer wg.Done()

			baseNamePerSpec := fmt.Sprintf("%s-%d", baseName, index)
			pathUsedAsHome, err := ctx.prepareUserHome(baseNamePerSpec, ticket.Sources)
			if err != nil {
				m.Lock()
				defer m.Unlock()

				errs[index] = err
				return
			}
			log.Printf("basePath => %s\n", pathUsedAsHome)

			// execute
			if err := ctx.execSpec(pathUsedAsHome, index, execSpec, callback); err != nil {
				m.Lock()
				defer m.Unlock()

				errs[index] = err
				return
			}
		}(index_, execSpec_)
	}
	wg.Wait()

	if len(errs) == 0 {
		return nil
	} else {
		return fmt.Errorf("%v", errs)
	}
}

func (ctx *Context) execSpec(
	pathUsedAsHome string,
	mainIndex int,
	execSpec *ExecutionSpec,
	callback invokeResultRecieverCallback,
) error {
	// build
	if execSpec.IsBuildRequired() {
		if err := ctx.execBuild(pathUsedAsHome, mainIndex, execSpec.BuildInst, callback); err != nil {
			if err == buildFailedError {
				return nil // not an error
			} else {
				return err
			}
		}
	}

	// run
	if err := ctx.execManagedRun(pathUsedAsHome, mainIndex, execSpec.RunInsts, callback); err != nil {
		return err
	}

	return nil
}

// ========================================
// ========================================
func (ctx *Context) execBuild(
	pathUsedAsHome string,
	mainIndex int,
	buildInst *BuildInstruction,
	callback invokeResultRecieverCallback,
) error {
	log.Printf("$$$$$$$$$$ START build => %s\n", pathUsedAsHome)
	defer log.Printf("$$$$$$$$$$ FINISH build  => %s\n", pathUsedAsHome)

	// compile phase
	log.Printf("$$$$$$$$$$ START compile\n")
	if err := ctx.invokeCompileCommand(
		pathUsedAsHome,
		mainIndex,
		buildInst.CompileSetting,
		callback,
	); err != nil {
		if err == compileFailedError {
			return buildFailedError
		} else {
			return err
		}
	}

	// link phase :: if link command is separated, so call linking commands
	if buildInst.IsLinkIndependent() {
		if err := ctx.invokeLinkCommand(
			pathUsedAsHome,
			mainIndex,
			buildInst.LinkSetting,
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
	pathUsedAsHome string,
	mainIndex int,
	runInsts []*RunInstruction,
	callback invokeResultRecieverCallback,
) error {
	log.Printf("$$$$$$$$$$ START run => %s\n", pathUsedAsHome)
	defer log.Printf("$$$$$$$$$$ FINISH run => %s\n", pathUsedAsHome)

	m := new(sync.Mutex)
	errs := map[int]error{}

	wg := new(sync.WaitGroup)
	for index_, runInst_ := range runInsts {
		wg.Add(1)
		go func(index int, runInst *RunInstruction) {
			defer wg.Done()

			if err := ctx.invokeRunCommand(pathUsedAsHome, mainIndex, index, runInst, callback); err != nil {
				m.Lock()
				defer m.Unlock()

				errs[index] = err
				return
			}
		}(index_, runInst_)
	}
	wg.Wait()

	if len(errs) == 0 {
		return nil
	} else {
		return fmt.Errorf("%v", errs)
	}
}

// ========================================
// ========================================
func (ctx *Context) invokeCompileCommand(
	pathUsedAsHome string,
	mainIndex int,
	settings *ExecutionSetting,
	callback invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeCompileCommand")

	opts := &SandboxExecutionOption{
		Mounts: []MountOption{
			MountOption{
				HostPath:   pathUsedAsHome,
				GuestPath:  guestHome,
				IsReadOnly: false,
				DoChown:    true,
			},
			MountOption{
				HostPath:   ctx.sandboxExecutor.DefaultMountOption().HostPath,
				GuestPath:  ctx.sandboxExecutor.DefaultMountOption().GuestPath,
				IsReadOnly: true,
				DoChown:    false,
			},
		},
		GuestHomePath: guestHome,
		Limits: &ResourceLimit{
			Core:    0,                         // Process can NOT create CORE file
			Nofile:  512,                       // Process can open 512 files
			NProc:   30,                        // Process can create processes to 30
			MemLock: 1024,                      // Process can lock 1024 Bytes by mlock(2)
			CpuTime: settings.CpuTimeLimit,     // sec
			Memory:  settings.MemoryBytesLimit, // bytes
			FSize:   5 * 1024 * 1024,           // Process can writes a file only 5MiB
		},
		Args: settings.Args,
		Envs: settings.Envs,
	}

	f := func(output *StreamOutput) error {
		return callback(&StreamOutputResult{
			Mode:      CompileMode,
			MainIndex: mainIndex,
			SubIndex:  0,
			Output:    output,
		})
	}
	result, err := ctx.sandboxExecutor.Execute(opts, nil, f)
	log.Printf("Compile Executed >> %v / %v", result, err)
	if err != nil {
		return err
	}

	callback(&StreamExecutedResult{
		Mode:      CompileMode,
		MainIndex: mainIndex,
		SubIndex:  0,
		Result:    result,
	})

	if !result.IsSucceeded() {
		return compileFailedError
	} else {
		return nil
	}
}

func (ctx *Context) invokeLinkCommand(
	pathUsedAsHome string,
	mainIndex int,
	settings *ExecutionSetting,
	callback invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeLinkCommand")

	opts := &SandboxExecutionOption{
		Mounts: []MountOption{
			MountOption{
				HostPath:   pathUsedAsHome,
				GuestPath:  guestHome,
				IsReadOnly: false,
				DoChown:    true,
			},
			MountOption{
				HostPath:   ctx.sandboxExecutor.DefaultMountOption().HostPath,
				GuestPath:  ctx.sandboxExecutor.DefaultMountOption().GuestPath,
				IsReadOnly: true,
				DoChown:    false,
			},
		},
		GuestHomePath: guestHome,
		Limits: &ResourceLimit{
			Core:    0,                      // Process can NOT create CORE file
			Nofile:  512,                    // Process can open 512 files
			NProc:   30,                     // Process can create processes to 30
			MemLock: 1024,                   // Process can lock 1024 Bytes by mlock(2)
			CpuTime: 10,                     // 10 sec
			Memory:  2 * 1024 * 1024 * 1024, // 2GiB[fixed]
			FSize:   40 * 1024 * 1024,       // 40MiB[fixed]
		},
		Args: settings.Args,
		Envs: settings.Envs,
	}

	f := func(output *StreamOutput) error {
		return callback(&StreamOutputResult{
			Mode:      LinkMode,
			MainIndex: mainIndex,
			SubIndex:  0,
			Output:    output,
		})
	}
	result, err := ctx.sandboxExecutor.Execute(opts, nil, f)
	log.Printf("Link Executed >> %v / %v", result, err)
	if err != nil {
		return err
	}

	callback(&StreamExecutedResult{
		Mode:      LinkMode,
		MainIndex: mainIndex,
		SubIndex:  0,
		Result:    result,
	})

	if !result.IsSucceeded() {
		return linkFailedError
	} else {
		return nil
	}
}

func (ctx *Context) invokeRunCommand(
	pathUsedAsHome string,
	mainIndex int,
	subIndex int,
	runInst *RunInstruction,
	callback invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeRunInputCommand")

	stdin := runInst.Stdin
	settings := runInst.RunSetting

	var temp_stdin *os.File = nil
	if stdin != nil {
		log.Println("use stdin")

		f, err := ioutil.TempFile("", "torigoya-inputs-")
		if err != nil {
			return err
		}
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
				HostPath:   ctx.sandboxExecutor.DefaultMountOption().HostPath,
				GuestPath:  ctx.sandboxExecutor.DefaultMountOption().GuestPath,
				IsReadOnly: true,
				DoChown:    false,
			},
		},
		Copies: []CopyOption{ // NOTE: NOT "Mount", to run async
			CopyOption{
				HostPath:  pathUsedAsHome,
				GuestPath: guestHome,
			},
		},
		GuestHomePath: guestHome,
		Limits: &ResourceLimit{
			Core:    0,                         // Process can NOT create CORE file
			Nofile:  512,                       // Process can open 512 files
			NProc:   30,                        // Process can create processes to 30
			MemLock: 1024,                      // Process can lock 1024 Bytes by mlock(2)
			CpuTime: settings.CpuTimeLimit,     // sec
			Memory:  settings.MemoryBytesLimit, // bytes
			FSize:   1 * 1024 * 1024,           // Process can writes a file only 1MB
		},
		Args: settings.Args,
		Envs: settings.Envs,
	}

	f := func(output *StreamOutput) error {
		return callback(&StreamOutputResult{
			Mode:      RunMode,
			MainIndex: mainIndex,
			SubIndex:  subIndex,
			Output:    output,
		})
	}
	result, err := ctx.sandboxExecutor.Execute(opts, temp_stdin, f)
	log.Printf("Run Executed >> %v / %v", result, err)
	if err != nil {
		return err
	}

	log.Printf(">> %v", result)
	callback(&StreamExecutedResult{
		Mode:      RunMode,
		MainIndex: mainIndex,
		SubIndex:  subIndex,
		Result:    result,
	})
	log.Println("sandboxExecutor.Exit >> %v", result)

	return nil
}

// ========================================
// ========================================
func (ctx *Context) prepareUserHome(
	baseName string,
	sources []*SourceData,
) (string, error) {
	// unpack source codes
	sourceContents, err := convertSourcesToContents(sources)
	if err != nil {
		return "", err
	}

	t := time.Now()
	const timeLayout = "2006-01-02-15-04-05.000"
	baseNameTmp := fmt.Sprintf("%s_%s", baseName, t.Format(timeLayout))
	userHomeDirPath, err := ctx.createMultipleTargets(
		baseNameTmp,
		sourceContents,
	)
	if err != nil {
		return "", fmt.Errorf("cannot create multi targets : %v", err.Error())
	}

	return userHomeDirPath, nil
}
