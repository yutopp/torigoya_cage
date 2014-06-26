//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

// +build linux

package torigoya

// #include <sys/resource.h>
import "C"

import(
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"
	"os"
	"os/exec"
	"syscall"
)


//
type ResourceLimit struct {
	CPU		uint64
	AS		uint64
	FSize	uint64
}


//
var errorSequence = []byte{ 0x0d, 0x0e, 0x0a, 0x0d }


//
func managedExec(
	rl *ResourceLimit,
	p *BridgePipes,
	args []string,
	envs map[string]string,
) (*ExecutedResult, error) {
	// make a pipe for error reports
	error_pipe, err := makePipe()
	if err != nil { return nil, err }
	defer error_pipe.Close()

	// fork process!
	pid, err := fork()
	if err != nil {
		return nil, err;
	}
	if pid == 0 {
		// child process
		child(rl, p, *error_pipe, args, envs)
		return nil, nil

	} else {
		// parent process

		//
		syscall.Close(error_pipe.WriteFd)

		//
		syscall.Close(p.Stdout.WriteFd)
		syscall.Close(p.Stderr.WriteFd)

		//
		process, err := os.FindProcess(pid)
		if err != nil {
			return nil, err;
		}

		// parent process
		wait_pid_chan := make(chan *os.ProcessState)
		go func() {
			ps, _ := process.Wait()
			wait_pid_chan <- ps
		}()

		//
		select {
		case ps := <-wait_pid_chan:
			usage, ok := ps.SysUsage().(*syscall.Rusage)
			if !ok {
				log.Fatal("akann")
			}
			fmt.Printf("%v\n", usage)

			// usage.Maxrss -> Amount of memory usage (KB)

			// error check sequence
			error_buf := make([]byte, 128)
			error_len, _ := syscall.Read(error_pipe.ReadFd, error_buf)
			if error_len < len(errorSequence) {
				// execution was succeeded
				return &ExecutedResult{
					IsSystemFailed: false,
				}, nil

			} else {
				// execution was failed
				if bytes.Equal(error_buf[:len(errorSequence)], errorSequence) {
					error_log := string(error_buf[4:error_len])
					for {
						size, err := syscall.Read(error_pipe.ReadFd, error_buf)
						if err != nil || size == 0 {
							break
						}

						error_log += string(error_buf[:size])
					}

					return nil, errors.New(error_log)

				} else {
					// invalid error byte sequence
					return nil, errors.New("invalid error byte sequence")
				}
			}

		case <-time.After(time.Duration(rl.CPU * 2) * time.Second):
			// timeout(e.g. when process uses sleep a lot)
			return nil, errors.New("TLE")
		}
	}
}


func child(
	rl *ResourceLimit,
	p *BridgePipes,
	error_pipe Pipe,
	args []string,
	envs map[string]string,
) {
	// if called this function, child process is failed to execute
	defer func() {
		// mark failed result
		syscall.Close(error_pipe.ReadFd)
		syscall.Write(error_pipe.WriteFd, errorSequence)		// write error sequence
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				syscall.Write(error_pipe.WriteFd, []byte(err.Error()))	// write panic sentence
			}
		}
		syscall.Close(error_pipe.WriteFd)

		// exit
		os.Exit(-1)
	}()

	print("aaaaaaaa ============= child")

	//
	setLimit(C.RLIMIT_CORE, 0)			// Process can NOT create CORE file
	setLimit(C.RLIMIT_NOFILE, 1024)		// Process can open 1024 files
	setLimit(C.RLIMIT_NPROC, 20)		// Process can create 20 processes
	setLimit(C.RLIMIT_MEMLOCK, 1024)	// Process can lock 1024 Bytes by mlock(2)

	setLimit(C.RLIMIT_CPU, rl.CPU)		// CPU can be used only cpu_limit_time(sec)
	setLimit(C.RLIMIT_AS, rl.AS)		// Memory can be used only memory_limit_bytes [be careful!]
	setLimit(C.RLIMIT_FSIZE, rl.FSize)	// Process can writes a file only 512 KBytes

	// TODO: stdin

	// redirect stdout
	if err := syscall.Close(p.Stdout.ReadFd); err != nil { panic(err) }
	if err := syscall.Dup2(p.Stdout.WriteFd, 1); err != nil { panic(err) }
	if err := syscall.Close(p.Stdout.WriteFd); err != nil { panic(err) }

	// redirect stderr
	if err := syscall.Close(p.Stderr.ReadFd); err != nil { panic(err) }
	if err := syscall.Dup2(p.Stderr.WriteFd, 2); err != nil { panic(err) }
	if err := syscall.Close(p.Stderr.WriteFd); err != nil { panic(err) }

	// set PATH env
	if path, ok := envs["PATH"]; ok {
		if err := os.Setenv("PATH", path); err != nil {
			panic(err)
		}
	}

	//
	if len(args) < 1 {
		panic(errors.New("args must contain at least one element"))
	}
	command := args[0]	//
	exec_path, err := exec.LookPath(command)
	if err != nil {
		panic(err)
	}

	//
	var env_list []string
	for k, v := range envs {
		env_list = append(env_list, k + "=" + v)
	}

	// exec!!
	err = syscall.Exec(exec_path, args, env_list);

	panic(errors.New(fmt.Sprintf("unreachable : " + err.Error())))
}


func fork() (int, error) {
	syscall.ForkLock.Lock()
	pid, _, err := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	syscall.ForkLock.Unlock()
	if err != 0 {
		return -1, err
	}
	return int(pid), nil
}


func setLimit(resource int, value uint64) {
	//
	if err := syscall.Setrlimit(resource, &syscall.Rlimit{value, value}); err != nil {
		panic(err)
	}
}
