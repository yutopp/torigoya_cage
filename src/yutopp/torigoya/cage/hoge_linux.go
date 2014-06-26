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


func (bm *BridgeMessage) invokeProcessCloner(
	cloner_dir		string,
) error {
	return invokeProcessClonerBase(cloner_dir, "process_cloner", bm)
}

//
func invokeProcessClonerBase(
	cloner_dir		string,
	cloner_name		string,
	bm				*BridgeMessage,
) error {
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
	if err != nil { return err }
	defer stdout_pipe.Close()

	stderr_pipe, err := makePipe()
	if err != nil { return err }
	defer stderr_pipe.Close()

	result_pipe, err := makePipe()
	if err != nil { return err }
	defer result_pipe.Close()

	// update pipe data to message
	bm.Pipes = &BridgePipes{
		Stdout: stdout_pipe.CopyForClone(),
		Stderr: stderr_pipe.CopyForClone(),
		Result: result_pipe.CopyForClone(),
	}
	//stdout_pipe.CloseWrite()
	//stderr_pipe.CloseWrite()
	//result_pipe.CloseWrite()

	//
	content_string, err := bm.Encode()
	if err != nil {
		return err
	}

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
		return err
	}

	// wait for exit process
	process_wait_c := make(chan error)
	go func() {
		process_wait_c <- cmd.Wait()
	}()

	// read stdout/stderr
	stdout_c := make(chan error)
	go readPipeAsync(stdout_pipe.ReadFd, stdout_c)
	stderr_c := make(chan error)
	go readPipeAsync(stderr_pipe.ReadFd, stderr_c)

	// wait for finishing subprocess
	select {
	case err := <-process_wait_c:
		// subprocess has been finished
		log.Printf("MYAN %v", err)
		if err != nil {
			return err
		}

		log.Printf("?? %v", cmd.ProcessState.Success())
		if !cmd.ProcessState.Success() {
			return errors.New("Process finished with failed state")
		}

		result_pipe.CloseWrite()
		result_buf, _ := readPipe(result_pipe.ReadFd)
		result, err := DecodeExecuteResult(result_buf)
		log.Printf("??RESULT!!!!!!! %v / %v", result, err)

	case <-time.After(500 * time.Second):
		// TODO: fix
		// will blocking( wait for response at least 500 seconds )
		log.Println("TIMEOUT")
		return errors.New("Process timeouted")
	}

	stdout_pipe.Close()
	stderr_pipe.Close()
	result_pipe.Close()

	return nil
}

func readPipeAsync(fd int, cs chan<- error) {
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
		}
		_ = size
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

func convertSourceToContent(
	sources []*SourceData,
) ([]TextContent, error) {
	source_contents := make([]TextContent, len(sources))

	//
	for i, s := range sources {
		data, err := func() ([]byte, error) {
			return s.Data, nil
		}()
		if err != nil {
			return nil, err
		}

		// collect file names
		source_contents[i] = TextContent{
			Name: s.Name,
			Data: data,
		}
	}

	return source_contents, nil
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
	input				*SourceData
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


//
func (ctx *Context) invokeBuild(
	base_name			string,
	sources				[]*SourceData,
	proc_profile		*ProcProfile,
	build_inst			*BuildInstruction,
	run_inst			*RunInstruction,
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
		)
	})

	return err
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
) error {
	log.Println(">> called invokeBuildCommand")

	// unpack source codes
	source_contents, err := convertSourceToContent(sources)
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
		message.invokeProcessCloner(bin_base_path)


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
			message.invokeProcessCloner(bin_base_path)
		}
	}
/*
	//
	user_dir_path, input_paths, err := ctx.reassignTarget(base_name, group_id, func(base_directory_name string) ([]string, error) {
		path, err := ctx.createInput(base_directory_name, group_id, stdin)
		if err != nil { return nil, err }
		return []string{ path }, nil
	})
	if err != nil {
		t.Errorf(err.Error())
		return
	}
*/

	return nil
}
