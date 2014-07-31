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
	"os/exec"
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

//
func invokeProcessClonerBase(
	cloner_dir		string,
	cloner_name		string,
	bm				*BridgeMessage,
	output_stream	chan<-StreamOutput,
) (*ExecutedResult, error) {
	cloner_path := filepath.Join(cloner_dir, cloner_name)
	log.Printf("Cloner path: %s", cloner_path)

	callback_path := filepath.Join(cloner_dir, "cage.callback")

	// init default value
	if bm == nil {
		bm = &BridgeMessage{}
	}

	// TODO: close on exec
	// pipe for
	stdout_pipe, err := makePipe()
	if err != nil { return nil, err }
	defer stdout_pipe.Close()

	stderr_pipe, err := makePipe()
	if err != nil { return nil, err }
	defer stderr_pipe.Close()

	result_pipe, err := makePipe()
	if err != nil { return nil, err }
	defer result_pipe.Close()

	// update pipe data to message
	bm.Pipes = &BridgePipes{
		Stdout: stdout_pipe.CopyForClone(),
		Stderr: stderr_pipe.CopyForClone(),
		Result: result_pipe.CopyForClone(),
	}

	//
	content_string, err := bm.Encode()
	if err != nil { return nil, err }

	//
	cmd := exec.Command(cloner_path)
	cmd.Env = []string{
		"callback_executable=" + callback_path,
		"packed_torigoya_content=" + content_string,
	}

	// debug...
	cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

	//cmd.Stdout = nil
    //cmd.Stderr = nil


	// Start Process
	// TODO: rewrite forc/exec to attach close-on-exec to pipe...
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
		return nil, err
	}

	// wait for exit process
	process_wait_c := make(chan error)
	go func() {
		process_wait_c <- cmd.Wait()
	}()

	// read stdout/stderr
	stdout_c := make(chan error)
	go readPipeAsync(stdout_pipe.ReadFd, stdout_c, StdoutFd, output_stream)
	stderr_c := make(chan error)
	go readPipeAsync(stderr_pipe.ReadFd, stderr_c, StderrFd, output_stream)

	//
	defer func() {
		// force close
		stdout_pipe.Close()
		stderr_pipe.Close()

		// block
		<- stdout_c
		<- stderr_c
	}()

	// wait for finishing subprocess
	select {
	case err := <-process_wait_c:
		// subprocess has been finished
		log.Printf("MYAN %v", err)
		if err != nil {
			return nil, err
		}

		log.Printf("?? %v", cmd.ProcessState.Success())
		if !cmd.ProcessState.Success() {
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
}

func readPipeAsync(fd int, cs chan<-error, output_fd OutFd, output_stream chan<-StreamOutput) {
	buffer := make([]byte, 1024)
	defer close(cs)

	for {
		size, err := syscall.Read(fd, buffer)
		if err != nil {
			cs <- err
			return
		}

		if size != 0 {
			log.Printf("= %d ==> %d", fd, size)
			log.Printf("= %d ==>\n%s\n<=====\n", fd, string(buffer[:size]))

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
	buffer := make([]byte, 1024)

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
