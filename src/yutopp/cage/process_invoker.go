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

	"time"
	"errors"
	"os"
	"syscall"
	"path/filepath"
)


//
const ReadLength = 8096

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
	output_stream	chan<-*StreamOutput,
	debug_tag		string,
) (*ExecutedResult, error) {
	log.Println(">> called invokeProcessCloner")

	return invokeProcessClonerBase(cloner_dir, "process_cloner", bm, output_stream, debug_tag)
}


func invokeProcessClonerBaseChild(
	cloner_dir		string,
	cloner_name		string,
	bm				*BridgeMessage,
	output_stream	chan<-*StreamOutput,
	debug_tag		string,
) (*ExecutedResult, error) {
	cloner_path := filepath.Join(cloner_dir, cloner_name)
	log.Printf("%s: Cloner path: %s", debug_tag, cloner_path)

	//
	new_stdout, err := bm.Pipes.Stdout.Dup()
	if err != nil { return nil, err }
	if err := bm.Pipes.Stdout.Close(); err != nil { return nil, err }

	new_stderr, err := bm.Pipes.Stderr.Dup()
	if err != nil { return nil, err }
	if err := bm.Pipes.Stderr.Close(); err != nil { return nil, err }

	new_result, err := bm.Pipes.Result.Dup()
	if err != nil { return nil, err }
	if err := bm.Pipes.Result.Close(); err != nil { return nil, err }

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
		"debug_tag=" + debug_tag,
	}

	//
	args := []string{
		cloner_name,
	}

	// exec!!
	err = syscall.Exec(cloner_path, args, envs);
	if err != nil {
		log.Printf("Error!!! %v", err)
	}

	// unreachable
	return nil, nil
}

//
func invokeProcessClonerBase(
	cloner_dir		string,
	cloner_name		string,
	bm				*BridgeMessage,
	output_stream	chan<-*StreamOutput,
	debug_tag		string,
) (*ExecutedResult, error) {
	// pipe for
	stdout_pipe, err := makePipeNonBlockingWithFlags(syscall.O_CLOEXEC)
	if err != nil { return nil, err }
	defer stdout_pipe.Close()

	stderr_pipe, err := makePipeNonBlockingWithFlags(syscall.O_CLOEXEC)
	if err != nil { return nil, err }
	defer stderr_pipe.Close()

	result_pipe, err := makePipe() //WithFlags(syscall.O_CLOEXEC)
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
	fmt.Printf("@@@@@@@@ after fork(pid): %s\n", pid, debug_tag)
	if err != nil {
		return nil, err
	}
	if pid == 0 {
		// child process
		invokeProcessClonerBaseChild(cloner_dir, cloner_name, bm, output_stream, debug_tag)

		// unreachable
		os.Exit(-1);
		return nil, nil

	} else {
		fmt.Printf("?? $ ?? fork: %s\n", debug_tag)

		// parent process
		process, err := os.FindProcess(pid)
		if err != nil {
			return nil, err
		}

		//
		stdout_pipe.CloseWrite()
		stderr_pipe.CloseWrite()
		result_pipe.CloseWrite()

		// parent process
		wait_pid_chan := make(chan *os.ProcessState)
		go func() {
			ps, _ := process.Wait()
			fmt.Printf("?? $ ?? FOUND: %s\n", debug_tag)
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
		force_quit := false

		//
		defer func() {
			log.Printf("%s: wait for closeing wait_pid_chan => %v\n", debug_tag, force_quit)
			close(wait_pid_chan)

			if force_quit {
				//force_close_out <- true
				//force_close_err <- true
			}

			// block

			// out
			log.Printf("%s: wait for recieving stdout_err\n", debug_tag)
			select {
			case err := <-stdout_err:
				if err != nil {
					log.Printf("??STDOUT ERROR: %v\n", err)
				}

			default:
				// not finished...
				force_close_out <- true
			}
			log.Printf("%s: wait for closeing stdout_err\n", debug_tag)
			close(force_close_out)

			// err
			log.Printf("%s: wait for recieving stderr_err\n", debug_tag)
			select {
			case err := <-stderr_err:
				if err != nil {
					log.Printf("??STDERR ERROR: %v\n", err)
				}

			default:
				// not finished...
				force_close_err <- true
			}
			log.Printf("%s: wait for closeing stderr_err\n", debug_tag)
			close(force_close_err)

			//
			log.Printf("%s: closed!\n", debug_tag)
		}()

		// wait for finishing subprocess
		select {
		case ps := <-wait_pid_chan:
			// subprocess has been finished
			log.Printf("!! CHILD PROCESS IS FINISHED %v", ps)

			if !ps.Success() {
				return nil, errors.New("Process finished with failed state")
			}

			result_buf_ch := make(chan []byte)
			result_err_ch := make(chan error)
			go func() {
				log.Printf("waiting a result...\n")
				result_buf, err := readPipe(result_pipe.ReadFd)
				if err != nil {
					result_err_ch <- err
					return
				}

				log.Printf("getting a result...\n")
				result_buf_ch <- result_buf
				log.Printf("got a result...\n")
			}()

			select {
			case result_buf := <-result_buf_ch:
				result, err := DecodeExecuteResult(result_buf)
				if err != nil { return nil, err }

				log.Printf("??RESULT!!!!!!! : err => %v", err)
				log.Printf("  => sec          : %v", result.UsedCPUTimeSec)
				log.Printf("  => mem          : %v", result.UsedMemoryBytes)
				log.Printf("  => signal       : %v", result.Signal)
				log.Printf("  => return code  : %v", result.ReturnCode)
				log.Printf("  => command      : %v", result.CommandLine)
				log.Printf("  => status       : %v", result.Status)
				log.Printf("  => system error : %v", result.SystemErrorMessage)

				if result.Status == 5 {
					force_quit = true
				}

				return result, err

			case result_err := <-result_err_ch:
				return nil, result_err

			case <-time.After(time.Second * 1):
				return nil, errors.New("Timeout, failed to get a result...")
			}

		case <-time.After(30 * time.Second):
			// TODO: fix
			// will blocking( wait for response at least 500 seconds )
			log.Printf("%s: TIMEOUT\n", debug_tag)
			return nil, errors.New("Process timeouted")
		}
	} // if pid
}

func readPipeAsync(
	fd int,
	cs chan<-error,
	output_fd OutFd,
	force_close_ch <-chan bool,
	output_stream chan<-*StreamOutput,
) {
	buffer := make([]byte, ReadLength)
	force_close_f := false

	for {
		select {
		case <-force_close_ch:
			force_close_f = true
		default:
		}

		size, err := syscall.Read(fd, buffer)
		if err != nil {
			if err == syscall.EAGAIN {
				continue
			} else {
				if ! force_close_f {
					cs <- err
				}

				output_stream <- nil
				return
			}
		}
		if size <= 0 && force_close_f {
			output_stream <- nil
			return
		}

		if size > 0 {
			copied := make([]byte, size)
			copy(copied, buffer[:size])

			output_stream <- &StreamOutput{
				Fd: output_fd,
				Buffer: copied,
			}
		}

		//
		// time.Sleep(1 * time.Millisecond)
	}

	// cs <- nil
}

func readPipe(fd int) (result []byte, err error) {
	buffer := make([]byte, ReadLength)

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
