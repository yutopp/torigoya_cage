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
	"fmt"
	"log"
	"strconv"
	"time"
	"errors"
	"io"
	"os"
	"os/user"
	"os/exec"
	"syscall"
	"path/filepath"

	"encoding/base64"
	"github.com/ugorji/go/codec"
)

type Context struct {
	basePath		string

	hostUser		*user.User

	sandboxDir		string
	homeDir			string
	jailedUserDir	string
}


func InitContext(base_path string) (*Context, error) {
	//
	sandbox_dir := "/tmp/sandbox"

	//
	host_user_name := "yutopp"
	host_user, err := user.Lookup(host_user_name)
	if err != nil {
		return nil, err
	}

	// In posix, Uid only contains numbers
	host_user_id, _ := strconv.Atoi(host_user.Uid)

	// create SANDBOX Directory, if not existed
	if !fileExists(sandbox_dir) {
		err := os.Mkdir(sandbox_dir, os.ModeDir | 0700)
		if err != nil {
			panic(fmt.Sprintf("Couldn't create directory %s", sandbox_dir))
		}

		if err := filepath.Walk(sandbox_dir, func(path string, info os.FileInfo, err error) error {
			if err != nil { return err }
			// r-x/---/---
			err = guardPath(path, host_user_id, host_user_id, 0500)
			return err
		}); err != nil {
			panic(fmt.Sprintf("Couldn't create directory %s", sandbox_dir))
		}
	}

	return &Context{
		basePath:			base_path,
		hostUser:			host_user,
		sandboxDir:			sandbox_dir,
		homeDir:			"home",
		jailedUserDir:		"home/torigoya",
	}, nil
}

func F() int {

	return 42
}

func expectRoot() {
	if os.Geteuid() != 0 {
		panic("run this program as root")
	}
}

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
		Stdout: stdout_pipe,
		Stderr: stderr_pipe,
		Result: result_pipe,
	}

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
	go readPipe(stdout_pipe.ReadFd, stdout_c)
	stderr_c := make(chan error)
	go readPipe(stderr_pipe.ReadFd, stderr_c)

	// wait for finishing subprocess
	select {
	case err := <-process_wait_c:
		// subprocess has been finished
		log.Println("MYAN")
		if err != nil {
			return err
		}

		log.Printf("?? %d", cmd.ProcessState.Success())
		if !cmd.ProcessState.Success() {
			return errors.New("Process finished with failed state")
		}

	case <-time.After(300 * time.Second):
		// will blocking( wait for response at least 300 seconds )

		log.Println("TIMEOUT")
	}
	stdout_pipe.Close()
	stderr_pipe.Close()

	return nil
}

func readPipe(fd int, cs chan<- error) {
	buffer := make([]byte, 1024)
	defer close(cs)

	for {
		size, err := syscall.Read(fd, buffer)
		if err != nil {
			if err == io.EOF {
				log.Printf("= EOF!!")
				break;
			} else {
				log.Printf("= MOUDAME!")
			}

			cs <- err
			return;
		}

		if size != 0 {
			log.Printf("= %d ==> %d", fd, size)
			log.Printf("= %d ==>\n%s\n<=====\n", fd, string(buffer[0:size]))
		}
		_ = size
	}

	cs <- nil
}
















type ProcTarget struct {
	Id		int
	Version	string
}


// For source codes, inputs
type SourceData struct {
	Name			string
	Data			[]byte
	IsCompressed	bool
}

func convertSourceToContent(
	sources []SourceData,
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


//
type ExecutionSetting struct {
	CommandLine			string
	StructuredCommand	[][]string
	CpuTimeLimit		uint64
	MemoryBytesLimit	uint64
}

type BuildInstruction struct {
	CompileSetting		*ExecutionSetting
	LinkSetting			*ExecutionSetting
}

type RunInstruction struct {
	Inputs				[]struct{ input *SourceData; setting *ExecutionSetting }
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
type BridgePipes struct {
	Stdout, Stderr, Result	*Pipe
}


//
type BridgeMessage struct {
	ChrootPath			string
	JailedUserHomePath	string
	JailedUser			*JailedUserInfo
	Pipes				*BridgePipes
	Message				ExecMessage
	IsReboot			bool
}

func (bm *BridgeMessage) Encode() (string, error) {
	var msgpack_bytes []byte
	enc := codec.NewEncoderBytes(&msgpack_bytes, &msgPackHandler)
	if err := enc.Encode(*bm); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(msgpack_bytes), nil
}

func DecodeBridgeMessage(base string) (*BridgeMessage, error) {
	decoded_bytes, err := base64.StdEncoding.DecodeString(base)
	if err != nil {
		return nil, err
	}

	bm := &BridgeMessage{}
	dec := codec.NewDecoderBytes(decoded_bytes, &msgPackHandler)
	if err := dec.Decode(bm); err != nil {
		return nil, err
	}

	return bm, nil
}


//
func (bm *BridgeMessage) Exec() error {
	m := bm.Message

	return func() error {
		switch m.Mode {
		case CompileMode:
			return bm.compile()
		case LinkMode:
			return bm.link()
		case RunMode:
			return bm.run()
		default:
			return errors.New("Invalid mode")
		}
	}()
}

//
func (bm *BridgeMessage) compile() error {
	exec_message := bm.Message

	proc_profile := exec_message.Profile
	stdin_file_path := exec_message.StdinFilePath
	exec_setting := exec_message.Setting

	// arguments
	args, err := proc_profile.Compile.MakeCompleteArgs(
		exec_setting.CommandLine,
		exec_setting.StructuredCommand,
	)
	if err != nil {
		return err
	}

	//
	res_limit := &ResourceLimit{
		CPU: exec_setting.CpuTimeLimit,		// CPU limit(sec)
		AS: exec_setting.MemoryBytesLimit,	// Memory limit(bytes)
		FSize: 5 * 1024 * 1024,				// Process can writes a file only 5 MBytes
	}

	_ = stdin_file_path

	managedExec(res_limit, bm.Pipes, args, map[string]string{"PATH": "/bin"})

	return nil
}

func (bm *BridgeMessage) link() error {
	return nil
}

func (bm *BridgeMessage) run() error {
	return nil
}



func (ctx *Context) invokeBuild(
	base_name			string,
	sources				[]SourceData,
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
	sources				[]SourceData,
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
