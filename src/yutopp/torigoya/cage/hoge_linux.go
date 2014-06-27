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


type OutFd		int
const (
	StdoutFd = OutFd(0)
	StderrFd = OutFd(1)
)
type StreamOutput struct {
	Fd			OutFd
	Buffer		[]byte
}

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

			output_stream <- StreamOutput{
				Fd: output_fd,
				Buffer: buffer[:size],
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




// ========================================
// For source codes, stdins
type SourceData struct {
	Name			string
	Data			[]byte
	IsCompressed	bool
}

func convertSourcesToContents(
	sources []*SourceData,
) (source_contents []*TextContent, err error) {
	source_contents = make([]*TextContent, len(sources))

	//
	for i, s := range sources {
		// collect file names
		source_contents[i], err = convertSourceToContent(s)
		if err != nil { return nil,err }
	}

	return source_contents, nil
}

func convertSourceToContent(
	source *SourceData,
) (*TextContent, error) {
	data, err := func() ([]byte, error) {
		// TODO: check that data is compressed
		return source.Data, nil
	}()
	if err != nil {
		return nil, err
	}

	return &TextContent{
		Name: source.Name,
		Data: data,
	}, nil
}


// ========================================
type ExecutionSetting struct {
	CommandLine			string
	StructuredCommand	[][]string
	CpuTimeLimit		uint64
	MemoryBytesLimit	uint64
}


// ========================================
type BuildInstruction struct {
	CompileSetting		*ExecutionSetting
	LinkSetting			*ExecutionSetting
}


// ========================================
type Input struct{
	stdin				*SourceData
	setting				*ExecutionSetting
}


// ========================================
type RunInstruction struct {
	Inputs				[]Input
}


// ========================================
type Ticket struct {
	BaseName		string
	ProcId			uint64
	ProcVersion		string
	Sources			[]*SourceData
	BuildInst		*BuildInstruction
	RunInst			*RunInstruction
}


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


// ========================================
func (ctx *Context) ExecTicket(
	ticker				*Ticket,
) error {
	return nil
}

type StreamOutputResult struct {
	Mode		int
	Index		int
	Output		StreamOutput
}

type invokeResultRecieverCallback		func(interface{})

//
func (ctx *Context) invokeBuild(
	base_name			string,
	sources				[]*SourceData,
	proc_profile		*ProcProfile,
	build_inst			*BuildInstruction,
	callback			invokeResultRecieverCallback,
) error {
	//
	user_dir_path := ctx.makeUserDirName(base_name)
	user_home_path := ctx.jailedUserDir
	bin_base_path := filepath.Join(ctx.basePath, "bin")

	//
	err := runAsManagedUser(func(jailed_user *JailedUserInfo) error {
		return ctx.invokeBuildCommand(
			user_dir_path,
			user_home_path,
			bin_base_path,
			jailed_user,

			base_name,
			sources,
			proc_profile,
			build_inst,
			callback,
		)
	})

	return err
}

func (ctx *Context) invokeRunCommand(
	user_dir_path		string,
	user_home_path		string,
	bin_base_path		string,
	jailed_user			*JailedUserInfo,

	base_name			string,
	proc_profile		*ProcProfile,
	run_inst			*RunInstruction,
	callback			invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeRunCommand")

	// ========================================
	for i, input := range run_inst.Inputs {
		// TODO: async
		err := runAsManagedUser(func(jailed_user *JailedUserInfo) error {
			return ctx.invokeRunInputCommandBase(
				user_dir_path,
				user_home_path,
				bin_base_path,
				jailed_user,

				base_name,
				proc_profile,
				&input,
				callback,
			)
		})

		_ = i
		_ = err
	}

	return nil
}




func (ctx *Context) invokeBuildCommand(
	user_dir_path		string,
	user_home_path		string,
	bin_base_path		string,
	jailed_user			*JailedUserInfo,

	base_name			string,
	sources				[]*SourceData,
	proc_profile		*ProcProfile,
	build_inst			*BuildInstruction,
	callback			invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeBuildCommand")

	// unpack source codes
	source_contents, err := convertSourcesToContents(sources)
	if err != nil {
		return err
	}

	//
	_, err = ctx.createMultipleTargets(base_name, jailed_user.GroupId, source_contents)
	if err != nil {
		return errors.New("couldn't create multi target : " + err.Error());
	}

	//
	if proc_profile.IsBuildRequired {
		if build_inst == nil { return errors.New("compile_dataset is nil") }

		if build_inst.CompileSetting == nil { return errors.New("build_inst.CompileSetting is nil") }
		message := BridgeMessage{
			ChrootPath: user_dir_path,
			JailedUserHomePath: user_home_path,
			JailedUser: jailed_user,
			Message: ExecMessage{
				Profile: proc_profile,
				Setting: build_inst.CompileSetting,
				Mode: CompileMode,
			},
			IsReboot: false,
		}

		//
		build_output_stream := make(chan StreamOutput)
		go func() {
			for out := range build_output_stream {
				log.Printf("%V\n", out)
/*
				if callback != nil {
					callback(StreamOutputResult{
						Mode: CompileMode,
						Index: 0,
						Output: out,
					})
				}
*/
			}
		}()

		//
		result, err := message.invokeProcessCloner(bin_base_path, build_output_stream)

		//
		close(build_output_stream)
		if err != nil { return err }
		if callback != nil { callback(result) }

		// if build is failed, do NOT linkng...
		if result.IsFailed() {
			return nil
		}


		if proc_profile.IsLinkIndependent {
			// link command is separated, so call linking commands
			if build_inst.LinkSetting == nil { return errors.New("build_inst.LinkSetting is nil") }
			message := BridgeMessage{
				ChrootPath: user_dir_path,
				JailedUserHomePath: user_home_path,
				JailedUser: jailed_user,
				Message: ExecMessage{
					Profile: proc_profile,
					Setting: build_inst.LinkSetting,
					Mode: LinkMode,
				},
				IsReboot: true,		// mark as to reinvoke cloner
			}

			//
			link_output_stream := make(chan StreamOutput)
			go func() {
				for out := range link_output_stream {
					if callback != nil {
						callback(StreamOutputResult{
							Mode: LinkMode,
							Index: 0,
							Output: out,
						})
					}
				}
			}()

			//
			result, err := message.invokeProcessCloner(bin_base_path, link_output_stream)

			//
			close(link_output_stream)
			if err != nil { return err }
			if callback != nil { callback(result) }
		}
	}

	return nil
}


func (ctx *Context) invokeRunInputCommandBase(
	user_dir_path		string,
	user_home_path		string,
	bin_base_path		string,
	jailed_user			*JailedUserInfo,

	base_name			string,
	proc_profile		*ProcProfile,
	input				*Input,
	callback			invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeRunInputCommand")

	// TODO: add lock
	// reassign base files to new user
	user_dir_path, input_paths, err := ctx.reassignTarget(
		base_name,
		jailed_user.GroupId,
		func(base_directory_name string) ([]string, error) {
			if input.stdin != nil {
				// stdin exists

				// unpack source codes
				stdin_content, err := convertSourceToContent(input.stdin)
				if err != nil { return nil, err }

				path, err := ctx.createInput(base_directory_name, jailed_user.GroupId, stdin_content)
				if err != nil { return nil, err }
				return []string{ path }, nil

			} else {
				// nothing to do
				return nil, nil
			}
		},
	)
	if err != nil { return err }

	//
	var stdin_path *string = nil
	if input.stdin != nil {
		if len(input_paths) != 1 { return errors.New("invalid stdin file") }
		stdin_path = &input_paths[0]
	}

	//
	message := BridgeMessage{
		ChrootPath: user_dir_path,
		JailedUserHomePath: user_home_path,
		JailedUser: jailed_user,
		Message: ExecMessage{
			Profile: proc_profile,
			StdinFilePath: stdin_path,
			Setting: input.setting,
			Mode: RunMode,
		},
		IsReboot: false,
	}
	run_output_stream := make(chan StreamOutput)
	result, err := message.invokeProcessCloner(bin_base_path, run_output_stream)
	close(run_output_stream)
	if err != nil { return err }
	_ = result

	return nil
}
