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
	"syscall"
	"path/filepath"
)


//
type OutFd		int
const (
	StdoutFd = OutFd(0)
	StderrFd = OutFd(1)
)
type StreamOutput struct {
	Fd			OutFd
	Buffer		[]byte
}

func (s *StreamOutput) ToTuple() []interface{} {
	return []interface{}{ s.Fd, s.Buffer }
}

//
func (bm *BridgeMessage) invokeProcessCloner(
	cloner_dir		string,
	output_stream	chan<-StreamOutput,
) (*ExecutedResult, error) {
	return invokeProcessClonerBase(cloner_dir, "process_cloner", bm, output_stream)
}


func invokeProcessClonerBaseChild(
	cloner_dir		string,
	cloner_name		string,
	bm				*BridgeMessage,
	output_stream	chan<-StreamOutput,
) (*ExecutedResult, error) {
	cloner_path := filepath.Join(cloner_dir, cloner_name)
	log.Printf("Cloner path: %s", cloner_path)

	//
	new_stdout, err := bm.Pipes.Stdout.Dup()
	if err != nil { return nil, err }

	new_stderr, err := bm.Pipes.Stderr.Dup()
	if err != nil { return nil, err }

	new_result, err := bm.Pipes.Result.Dup()
	if err != nil { return nil, err }

	//
	bm.Pipes = &BridgePipes{
		Stdout: new_stdout,
		Stderr: new_stderr,
		Result: new_result,
	}

	callback_path := filepath.Join(cloner_dir, "cage.callback")

	//
	content_string, err := bm.Encode()
	if err != nil { return nil, err }

	//
	envs := []string{
		"callback_executable=" + callback_path,
		"packed_torigoya_content=" + content_string,
	}

	//
	args := []string{
		cloner_name,
	}

	// exec!!
	err = syscall.Exec(callback_path, args, envs);
	if err != nil {
		log.Printf("Error!!! %v", err)
	}

	return nil, nil
}

//
func invokeProcessClonerBase(
	cloner_dir		string,
	cloner_name		string,
	bm				*BridgeMessage,
	output_stream	chan<-StreamOutput,
) (*ExecutedResult, error) {
	// pipe for
	stdout_pipe, err := makePipeNonBlockingWithFlags(syscall.O_CLOEXEC)
	if err != nil { return nil, err }
	defer stdout_pipe.Close()

	stderr_pipe, err := makePipeNonBlockingWithFlags(syscall.O_CLOEXEC)
	if err != nil { return nil, err }
	defer stderr_pipe.Close()

	result_pipe, err := makePipeWithFlags(syscall.O_CLOEXEC)
	if err != nil { return nil, err }
	defer result_pipe.Close()

	// init default value
	if bm == nil {
		bm = &BridgeMessage{}
	}
	// update pipe data to message
	bm.Pipes = &BridgePipes{
		Stdout: stdout_pipe,
		Stderr: stderr_pipe,
		Result: result_pipe,
	}

	// fork process!
	pid, err := fork()
	if err != nil {
		return nil, err;
	}
	if pid == 0 {
		// child process
		invokeProcessClonerBaseChild(cloner_dir, cloner_name, bm, output_stream)

		// unreachable
		os.Exit(-1);
		return nil, nil

	} else {
		// parent process
		process, err := os.FindProcess(pid)
		if err != nil {
			return nil, err
		}

		//
		stdout_pipe.CloseWrite()
		stderr_pipe.CloseWrite()

		// parent process
		wait_pid_chan := make(chan *os.ProcessState)
		go func() {
			ps, _ := process.Wait()
			wait_pid_chan <- ps
		}()

		// read stdout/stderr
		force_close_out := make(chan bool)
		stdout_err := make(chan error)
		go readPipeAsync(stdout_pipe.ReadFd, stdout_err, StdoutFd, force_close_out, output_stream)
		force_close_err := make(chan bool)
		stderr_err := make(chan error)
		go readPipeAsync(stderr_pipe.ReadFd, stderr_err, StderrFd, force_close_err, output_stream)

		//
		defer func() {
			close(wait_pid_chan)

			force_close_out <- true
			force_close_err <- true

			// block
			if err := <-stdout_err; err != nil {
				log.Printf("??STDOUT ERROR: %v\n", err)
			}
			close(force_close_out)

			if err := <-stderr_err; err != nil {
				log.Printf("??STDERR ERROR: %v\n", err)
			}
			close(force_close_err)
		}()

		// wait for finishing subprocess
		select {
		case ps := <-wait_pid_chan:
			// subprocess has been finished
			log.Printf("!! CHILD PROCESS IS FINISHED %v", ps)

			if !ps.Success() {
				return nil, errors.New("Process finished with failed state")
			}

			//
			result_pipe.CloseWrite()
			result_buf, _ := readPipe(result_pipe.ReadFd)
			result, err := DecodeExecuteResult(result_buf)
			log.Printf("??RESULT!!!!!!! %v / %v", result, err)

			return result, err

		case <-time.After(500 * time.Second):
			// TODO: fix
			// will blocking( wait for response at least 500 seconds )
			log.Println("TIMEOUT")
			return nil, errors.New("Process timeouted")
		}
	} // if pid






}

func readPipeAsync(
	fd int,
	cs chan<-error,
	output_fd OutFd,
	force_close_ch <-chan bool,
	output_stream chan<-StreamOutput,
) {
	buffer := make([]byte, 2048)
	defer close(cs)

	force_close_f := false

	for {
		select {
		case f := <-force_close_ch:
			force_close_f = f
		default:
		}

		size, err := syscall.Read(fd, buffer)
		//log.Printf("=================== %v / = %v", size, err)
		if err != nil {
			if err != syscall.EAGAIN {
				cs <- err
				return
			}
		}

		if size <= 0 {
			// not error, there is no data to read
			if force_close_f {
				cs <- nil
				return
			}

		} else {
			//log.Printf("= %d ==> %d", fd, size)
			//log.Printf("= %d ==>\n%s\n<=====\n", fd, string(buffer[:size]))

			//
			copied := make([]byte, size)
			copy(copied, buffer[:size])

			//
			output_stream <- StreamOutput{
				Fd: output_fd,
				Buffer: copied,
			}
		}
	}


	cs <- nil
}

func readPipe(fd int) (result []byte, err error) {
	buffer := make([]byte, 2048)

	for {
		size, err := syscall.Read(fd, buffer)
		if err != nil {
			break
		}

		if size != 0 {
			result = append(result, buffer[:size]...)
		} else {
			break
		}
	}

	return
}
