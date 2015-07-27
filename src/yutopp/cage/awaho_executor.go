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

	"time"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"encoding/json"
	"sync"
	"io"
)

const ReadLength = 8096

type awahoResult struct {
	Exited				bool
	ExitStatus			int
	Signaled			bool
	Signal				int

	SystemTimeMicroSec	float64
	UserTimeMicroSec	float64
	CpuTimeMicroSec		float64

	UsedMemoryBytes		uint64

	SystemErrorStatus	int
	SystemErrorMessage	string
}

type resultPair struct {
	result	*awahoResult
	err		error
}


type awahoSandboxExecutor struct {
	ExecutablePath		string
}

func MakeAwahoSandboxExecutor(
	executablePath		string,
) (*awahoSandboxExecutor, error) {
	cwd, err := os.Getwd()
	if err != nil { return nil, err }

	path := func() string {
		if filepath.IsAbs(executablePath) {
			return executablePath
		} else {
			return filepath.Join(cwd, executablePath)
		}
	}()

	return &awahoSandboxExecutor{
		ExecutablePath: path,
	}, nil
}

func (exec *awahoSandboxExecutor) Execute(
	opts		*SandboxExecutionOption,
	stdin_f		*os.File,					// nullable
	callback	ExecuteCallBackType,
) (*ExecutedResult, error) {
	log.Printf(">> AwahoSandboxExecutor::Execute path=%s\n", exec.ExecutablePath)

	// for stdout / CLOSE_EXEC
	stdout, stdout_dtor, err := makePipe()
	if err != nil { return nil, err }
	defer stdout_dtor()

	// for stderr / CLOSE_EXEC
	stderr, stderr_dtor, err := makePipe()
	if err != nil { return nil, err }
	defer stderr_dtor()

	// for result / CLOSE_EXEC
	result_p, result_p_dtor, err := makePipe()
	if err != nil { return nil, err }
	defer result_p_dtor()

	//
	args := []string{
		exec.ExecutablePath,
		"--start-guest-path", opts.GuestHomePath,
		"--pipe", "4:1",	// (stdout in sandbox)
		"--pipe", "5:2",	// (stderr in sandbox)
		"--result-fd", "6",	// result reciever
		"--core", strconv.FormatUint(opts.Limits.Core, 10),
		"--nofile", strconv.FormatUint(opts.Limits.Nofile, 10),
		"--nproc", strconv.FormatUint(opts.Limits.NProc, 10),
		"--memlock", strconv.FormatUint(opts.Limits.MemLock, 10),
		"--cputime", strconv.FormatUint(opts.Limits.CpuTime, 10),
		"--memory", strconv.FormatUint(opts.Limits.Memory, 10),
		"--fsize", strconv.FormatUint(opts.Limits.FSize, 10),
	}
	if stdin_f != nil {
		args = append(args, []string{
			"--pipe", "3:0",	// (stdin in sandbox)
		}...)
	}
	args = append(args, exec.makeMountOptions(opts)...)
	args = append(args, exec.makeCopyOptions(opts)...)
	args = append(args, exec.makeEnvOptions(opts)...)
	args = append(args, "--")
	args = append(args, opts.Args...)

	attr := os.ProcAttr{
		Files: []*os.File{
			nil,		// 0 (not be used)
			os.Stdout,	// 1
			os.Stderr,	// 2
			stdin_f,	// 3 (stdin in sandbox)
			stdout.w,	// 4 (stdout in sandbox)
			stderr.w,	// 5 (stderr in sandbox)
			result_p.w,	// 6 result reciever
		},
		Env: []string{},
	}

	// invoke sandbox executor
	process, err := os.StartProcess(exec.ExecutablePath, args, &attr)
	if err != nil {
		return nil, err
	}

	if err = stdout.w.Close(); err != nil { return nil, err }
	if err = stderr.w.Close(); err != nil { return nil, err }
	if err = result_p.w.Close(); err != nil { return nil, err }

	//
	stdout_read_err_ch := readPipeOutputAsync(stdout, StdoutFd, callback)
	stderr_read_err_ch := readPipeOutputAsync(stderr, StderrFd, callback)

	// blocking, wait for finish process
	ps, _ := process.Wait()
	log.Printf("=> process finished")

	if !ps.Success() {
		// if awaho finished with failed state, it denotes host error
		return nil, errors.New("Process finished with failed state")
	}

	// TODO: to async
	// read result
	result, err := func() (*awahoResult, error) {
		select {
		case res := <-readResultAsync(result_p):
			return res.result, err

		case <-time.After(time.Second * 2):
			return nil, errors.New("Timeout")
		}
	}()
	if err != nil {
		return nil, err
	}

	stdout_read_err := <-stdout_read_err_ch
	stderr_read_err := <-stderr_read_err_ch

	// make generic result
	executed_result := &ExecutedResult{
		Exited: result.Exited,
		ExitStatus: result.ExitStatus,
		Signaled: result.Signaled,
		Signal: result.Signal,

		UsedCPUTimeSec: result.CpuTimeMicroSec / 1e6,	// micro sec to sec
		UsedMemoryBytes: result.UsedMemoryBytes,

		SystemErrorStatus: result.SystemErrorStatus,
		SystemErrorMessage: result.SystemErrorMessage,
	}

	log.Printf("terminate", stdout_read_err, stderr_read_err, *result, *executed_result)

	return executed_result, nil
}


func (exec *awahoSandboxExecutor) makeMountOptions(
	opts	*SandboxExecutionOption,
) []string {
	if opts.Mounts == nil {
		return []string{}
	}

	xs := make([]string, len(opts.Mounts) * 2)
	for _, mount := range opts.Mounts {
		aux := func() string {
			if mount.IsReadOnly {
				return "ro"
			} else {
				return "rw"
			}
		}()

		xs = append(
			xs,
			[]string{
				"--mount", mount.HostPath + ":" + mount.GuestPath + ":" + aux,
			}...
		)
	}

	return xs
}

func (exec *awahoSandboxExecutor) makeCopyOptions(
	opts	*SandboxExecutionOption,
) []string {
	if opts.Copies == nil {
		return []string{}
	}

	xs := make([]string, len(opts.Copies) * 2)
	for _, copy := range opts.Copies {
		xs = append(
			xs,
			[]string{
				"--copy", copy.HostPath + ":" + copy.GuestPath,
			}...
		)
	}

	return xs
}

func (exec *awahoSandboxExecutor) makeEnvOptions(
	opts	*SandboxExecutionOption,
) []string {
	if opts.Envs == nil {
		return []string{}
	}

	xs := make([]string, len(opts.Envs) * 2)
	for _, env := range opts.Envs {
		xs = append(
			xs, []string{"--env", env}...
		)
	}

	return xs
}


type pipeFiles struct {
	r	*os.File		// io for read
	w	*os.File		// io for write
	m	*sync.Mutex		// mutex
}

func makePipe() (*pipeFiles, func(), error) {
	stdout_m := new(sync.Mutex)
	stdout_r, stdout_w, err := os.Pipe()
	if err != nil { return nil, nil, err }

	dtor := func() {
		stdout_m.Lock()
		stdout_r.Close()
		stdout_w.Close()
		stdout_m.Unlock()
	}

	return &pipeFiles{
		r: stdout_r,
		w: stdout_w,
		m: stdout_m,
	}, dtor, nil
}

func readPipeOutputAsync(
	pipe		*pipeFiles,
	fdAs		OutFd,
	callback	ExecuteCallBackType,
) <-chan error {
	ch := make(chan error)

	go func() {
		buffer := make([]byte, ReadLength)

		for {
			pipe.m.Lock()
			log.Printf("=> READREADREAD")
			size, err := pipe.r.Read(buffer)
			log.Printf("=> READREADREAD => size=%v err=%v", size, err)
			pipe.m.Unlock()
			if err != nil {
				if err == io.EOF {
					log.Printf("Terminate success fully")
					ch <- nil
					break
				} else {
					log.Printf("Failed to Read: %v", err)
					ch <- nil	// TODO: change to err
					break
				}
			}
			if size > 0 {
				copied := make([]byte, size)
				copy(copied, buffer[:size])

				callback(&StreamOutput{
					Fd: fdAs,
					Buffer: copied,
				})
			}

			log.Printf("=> %V", string(buffer[:size]))
		}
	}()

	return ch
}

func readResultAsync(
	pipe		*pipeFiles,
) <-chan resultPair {
	ch := make(chan resultPair)

	go func() {
		pipe.m.Lock()
		defer pipe.m.Unlock()

		dec := json.NewDecoder(pipe.r)

		var result_detail awahoResult
		if err := dec.Decode(&result_detail); err != nil {
			ch <- resultPair{ nil, err }
			return
		}

		ch <- resultPair{ &result_detail, nil }
	}()

	return ch
}
