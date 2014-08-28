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
	"strings"
	"errors"
	"fmt"
	"time"
	"os"
	"os/exec"
	"syscall"
	"log"
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
	rl					*ResourceLimit,
	p					*BridgePipes,
	args				[]string,
	envs				map[string]string,
	umask				int,
	stdin_file_path		*string,
) (*ExecutedResult, error) {
	// make a pipe for error reports
	error_pipe, err := makePipeCloseOnExec()
	if err != nil { return nil, err }
	defer error_pipe.Close()

	log.Printf(">> managedExec start\n")

	// fork process!
	pid, err := fork()
	if err != nil {
		return nil, err;
	}
	if pid == 0 {
		// !! call child process !!
		log.Printf(">> managedExec || child\n")

		managedExecChild(rl, p, *error_pipe, args, envs, umask, stdin_file_path)
		return nil, nil

	} else {
		// parent process
		log.Printf(">> managedExec || parent\n")

		//
		defer func() {
			syscall.Close(p.Stdout.WriteFd)
			syscall.Close(p.Stderr.WriteFd)
		}()

		//
		syscall.Close(error_pipe.WriteFd)

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
			// take status

			//
			usage, ok := ps.SysUsage().(*syscall.Rusage)
			if !ok {
				return nil, errors.New("failed to cast to *syscall.Rusage")
			}
			fmt.Printf("Usage %v\n", usage)

			// error check sequence
			error_buf := make([]byte, 128)
			error_len, _ := syscall.Read(error_pipe.ReadFd, error_buf)
			if error_len < len(errorSequence) {
				// execution was succeeded
				wait_status, ok := ps.Sys().(syscall.WaitStatus)
				if !ok {
					return nil, errors.New("failed to cast to syscall.WaitStatus")
				}

				// take signal
				signal := func() *syscall.Signal {
					switch {
					case wait_status.Signaled():
						s := wait_status.Signal()
						return &s
						/*
					case wait_status.Stopped():
						return wait_status.StopSignal()
*/
					default:
						return nil
					}
				}()

				// exit status
				return_code := wait_status.ExitStatus()

				// take status
				status := func() ExecutedStatus {
					if ps.Success() {
						return Passed
					} else {
						return Error
					}
				}()

				// CPU time
				user_time := usage.Utime
				system_time := usage.Stime

				cpu_time := float32(user_time.Nano()) / 1e9 + float32(system_time.Nano()) / 1e9

				// Memory usage
				// usage.Maxrss -> Amount of memory usage (KB)
				// TODO: fix it
				memory := uint64(usage.Maxrss * 1024)

				// make result
				return &ExecutedResult{
					UsedCPUTimeSec: cpu_time,
					UsedMemoryBytes: memory,
					Signal: signal,
					ReturnCode: return_code,
					CommandLine: strings.Join(args, " "),
					Status: status,
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

		case <-time.After(time.Duration(rl.CPU * 2 + 20) * time.Second):
			// timeout(e.g. when process uses sleep a lot)
			return nil, errors.New("TLE")
		}
	}
}


func managedExecChild(
	rl					*ResourceLimit,
	p					*BridgePipes,
	error_pipe			Pipe/*close on exec*/,
	args				[]string,
	envs				map[string]string,
	umask				int,
	stdin_file_path		*string,
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

	log.Printf("== Managed: child           (%v)\n", args)
	log.Printf("== Managed: envs            (%v)\n", envs)
	log.Printf("== Managed: CPU(sec)        (%v)\n", rl.CPU)
	log.Printf("== Managed: memory(byte)    (%v)\n", rl.AS)
	log.Printf("== Managed: fsize           (%v)\n", rl.FSize)

	//
 	setLimit(C.RLIMIT_CORE, 0)			// Process can NOT create CORE file
 	setLimit(C.RLIMIT_NOFILE, 512)		// Process can open 512 files
	setLimit(C.RLIMIT_NPROC, 30)		// Process can create processes to 30
 	setLimit(C.RLIMIT_MEMLOCK, 1024)	// Process can lock 1024 Bytes by mlock(2)
//
 	setLimit(C.RLIMIT_CPU, rl.CPU)		// CPU can be used only cpu_limit_time(sec)
 	setLimit(C.RLIMIT_AS, rl.AS)		// Memory can be used only memory_limit_bytes [be careful!]
 	setLimit(C.RLIMIT_FSIZE, rl.FSize)	// Process can writes a file only FSize Bytes

	//
	//syscall.Umask(umask)

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
	command := args[0]	// args[0] is program name
	exec_path, err := exec.LookPath(command)
	if err != nil {
		panic(err)
	}

	//
	var env_list []string
	for k, v := range envs {
		env_list = append(env_list, k + "=" + v)
	}

	log.Printf("managed exec :: syscall.Exec!\n")

	// close unused pipe
	if err := p.Result.Close(); err != nil { panic(err) }


	log.Printf("==================================================\n")
	out, err := exec.Command("/bin/ls", "-laR", "/home/torigoya").Output()
	if err != nil {
		log.Printf("error:: %s\n", err.Error())
	} else {
		log.Printf("passed:: %s\n", out)
	}


	// redirect stdin
	if stdin_file_path != nil {
		log.Printf("============= stdin (%v)\n", *stdin_file_path)
		file, err := os.Open(*stdin_file_path)	// read
		if err != nil { panic(err) }
		defer file.Close()

		//
		if err := syscall.Dup2(int(file.Fd()), 0); err != nil { panic(err) }
	}

	// redirect stdout
	if err := p.Stdout.CloseRead(); err != nil { panic(err) }
	if err := syscall.Dup2(p.Stdout.WriteFd, 1); err != nil { panic(err) }
	if err := p.Stdout.CloseWrite(); err != nil { panic(err) }

	// redirect stderr
	if err := p.Stderr.CloseRead(); err != nil { panic(err) }
	if err := syscall.Dup2(p.Stderr.WriteFd, 2); err != nil { panic(err) }
	if err := p.Stderr.CloseWrite(); err != nil { panic(err) }

	// ==========
	// exec!!
	err = syscall.Exec(exec_path, args, env_list);

	panic(errors.New(fmt.Sprintf("UNREACHABLE!! managedExecChild / failed to Exec. Error => " + err.Error())))
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
