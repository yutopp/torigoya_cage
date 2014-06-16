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

func fileExists(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil
}



func (ctx *Context) makeUserDirName(base_name string) string {
	return filepath.Join(ctx.sandboxDir, base_name)
}


//
type createTargetCallback func(*os.File) (error)

//
func (ctx *Context) createTarget(
	base_name string,
	managed_group_id int,
	source_file_name string,
	callback createTargetCallback,
) (string, error) {
	source_full_paths, err := ctx.createMultipleTarget(
		base_name,
		managed_group_id,
		[]string{source_file_name},
		[]createTargetCallback{callback},
	)
	if err != nil {
		return "", err
	}
	if len(source_full_paths) != 1 {
		return "", errors.New("???")
	}

	return source_full_paths[0], err
}

//
func (ctx *Context) createMultipleTarget(
	base_name string,
	managed_group_id int,
	source_file_names []string,
	callbacks []createTargetCallback,
) (source_full_paths []string, err error) {
	log.Println("called SekiseiRunnerNodeServer::create_target")

	//
	if len(source_file_names) != len(callbacks) {
		return nil, errors.New("Size of source_file_names and callbacks are different")
	}

	//
    expectRoot()

	fmt.Printf("Euid -> %d\n", os.Geteuid())

	// In posix, Uid only contains numbers
	host_user_id, _ := strconv.Atoi(ctx.hostUser.Uid)
	fmt.Printf("host uid: %s\n", ctx.hostUser.Uid)

	//
	if !fileExists(ctx.sandboxDir) {
		panic(fmt.Sprintf("directory %s is not existed", ctx.sandboxDir))
	}

	// ========================================
	//// create user directory

	//
	user_dir_path := ctx.makeUserDirName(base_name)

	//
	if fileExists(user_dir_path) {
		log.Printf("user directory %s is already existed, so remove them\n", user_dir_path)
		err := os.RemoveAll(user_dir_path)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Couldn't remove directory %s (%s)", user_dir_path, err))
		}
	}

	//
	if err := os.Mkdir(user_dir_path, os.ModeDir); err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't create directory %s (%s)", user_dir_path, err))
	}
	// host_user_id:host_user_id // r-x/r-x/---
	if err := guardPath(user_dir_path, host_user_id, managed_group_id, 0550); err != nil {
		return nil, err
	}

	// ========================================
	//// create user HOME directory

	// create /home
	user_home_base_path := filepath.Join(user_dir_path, ctx.homeDir)
	if err := os.Mkdir(user_home_base_path, os.ModeDir); err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't create directory %s (%s)", user_home_base_path, err))
	}
	// host_user_id:managed_group_id // r-x/---/---
	if err := guardPath(user_home_base_path, host_user_id, managed_group_id, 0500); err != nil {
		return nil, err
	}

	// create /home/torigoya
	user_home_path := filepath.Join(user_dir_path, ctx.jailedUserDir)
	if err := os.Mkdir(user_home_path, os.ModeDir); err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't create directory %s (%s)", user_home_path, err))
	}
	// host_user_id:managed_group_id // rwx/r-x/---
	if err := guardPath(user_home_path, host_user_id, managed_group_id, 0750); err != nil {
		return nil, err
	}

	// ========================================
	//// make source file
	source_full_paths = make([]string, len(source_file_names))
	for index, source_file_name := range source_file_names {
		source_full_path := filepath.Join(user_home_path, source_file_name)
		f, err := os.OpenFile(source_full_path, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return nil, err
		}
		defer func() {
			f.Close()
			log.Printf("-> %s\n", source_full_path)
			// host_user_id:managed_group_id // r--/---/---
			err = guardPath(source_full_path, host_user_id, managed_group_id, 0400)
		}()

		//
		err = callbacks[index](f)

		//
		source_full_paths[index] = source_full_path
	}

	return source_full_paths, err
}


//
type reassignTargetCallback func(string) (string, error)

//
func (ctx *Context) reassignTarget(
	base_name string,
	managed_group_id int,
	callback reassignTargetCallback,
) (user_dir_path string, input_path string, err error) {
	log.Println("called SekiseiRunnerNodeServer::reassign_target")

    expectRoot()

	// In posix, Uid only contains numbers
	host_user_id, _ := strconv.Atoi(ctx.hostUser.Uid)
	fmt.Printf("host uid: %s\n", ctx.hostUser.Uid)

	//
	user_dir_path = ctx.makeUserDirName(base_name)

	// delete directories exclude HOME
	dirs, err := filepath.Glob(user_dir_path + "/*")
	if err != nil {
		return "", "", err
	}
	for _, dir := range dirs {
		rel_dir, err := filepath.Rel(user_dir_path, dir)
		if err != nil {
			return "", "", err
		}

		if rel_dir != ctx.homeDir {
			err := os.RemoveAll(dir)
			if err != nil {
				return "", "", errors.New(fmt.Sprintf("Couldn't remove directory %s (%s)", dir, err))
			}
		}
	}

	// chmod /home // host_user_id:managed_group_id // r-x/r-x/---
	user_home_base_path := filepath.Join(user_dir_path, ctx.homeDir)
	if err := guardPath(user_home_base_path, host_user_id, managed_group_id, 0550); err != nil {
		return "", "", err
	}

	// chmod /home/torigoya
	user_home_path := filepath.Join(user_dir_path, ctx.jailedUserDir)
	// host_user_id:managed_group_id // rwx/---/---
	if err := guardPath(user_home_path, host_user_id, managed_group_id, 0700); err != nil {
		return "", "", err
	}

	// call user block
	input_path, err = callback(user_dir_path)
	if err != nil {
		return "", "", err
	}

	// host_user_id:managed_group_id // rwx/r-x/---
	if err := guardPath(user_home_path, host_user_id, managed_group_id, 0750); err != nil {
		return "", "", err
	}

	//
	err = filepath.Walk(user_home_path, func(path string, info os.FileInfo, err error) error {
		log.Println(path)
		if err != nil { return err }
		if err := os.Chown(path, host_user_id, managed_group_id); err != nil {
			return errors.New(fmt.Sprintf("Couldn't chown %s, %s", path, err.Error()))
		}
		return err
	})

	return user_dir_path, input_path, err
}


func (ctx *Context) createInput(
	base_dir_path string,
	managed_group_id int,
	stdin_name string,
	stdin_content string,
) (stdin_full_path string, err error) {
	log.Println("called SekiseiRunnerNodeServer::createInput")

    expectRoot()

	// In posix, Uid only contains numbers
	host_user_id, _ := strconv.Atoi(ctx.hostUser.Uid)
	fmt.Printf("host uid: %s\n", ctx.hostUser.Uid)

	//
	const inputs_dir_name = "stdin"
	inputs_dir_path := filepath.Join(base_dir_path, ctx.jailedUserDir, inputs_dir_name)

	//
	if !fileExists(inputs_dir_path) {
		err := os.Mkdir(inputs_dir_path, os.ModeDir)
		if err != nil {
			panic(fmt.Sprintf("Couldn't create directory %s", inputs_dir_path))
		}
	}
	// host_user_id:managed_group_id // rwx/---/---
	if err := guardPath(inputs_dir_path, host_user_id, managed_group_id, 0700); err != nil {
		return "", err
	}

	//
	stdin_full_path = filepath.Join(inputs_dir_path, stdin_name)
	f, err := os.OpenFile(stdin_full_path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return "", err
	}
	defer func() {
 		f.Close()
		// host_user_id:managed_group_id // r--/---/---
		err = guardPath(stdin_full_path, host_user_id, managed_group_id, 0400)
	}()
	if _, err := f.WriteString(stdin_content); err != nil {
		return "", err
	}

	// host_user_id:managed_group_id // r-x/---/---
	if err := guardPath(inputs_dir_path, host_user_id, managed_group_id, 0500); err != nil {
		return "", err
	}

	return stdin_full_path, err
}


// if runnable file(a.out, main.py, etc..) exist, return true
func (ctx *Context) isTargetCached(
	base_name string,
	target_name string,
) bool {
	expectRoot()

	user_dir_path := ctx.makeUserDirName(base_name)
	target_path := filepath.Join(user_dir_path, ctx.jailedUserDir, target_name)

	return fileExists(target_path)
}


func guardPath(file_path string, user_id int, group_id int, mode os.FileMode) error {
	if err := os.Chown(file_path, user_id, group_id); err != nil {
		return errors.New(fmt.Sprintf("Couldn't chown %s, %s", file_path, err.Error()))
	}
	if err := os.Chmod(file_path, mode); err != nil {
		return errors.New(fmt.Sprintf("Couldn't chmod %s, %s", file_path, err.Error()))
	}

	return nil
}


func (passing_info *BrigdeInfo) invokeProcessCloner(
	cloner_dir		string,
) error {
	return invokeProcessClonerBase(cloner_dir, "process_cloner", passing_info)
}

//
func invokeProcessClonerBase(
	cloner_dir		string,
	cloner_name		string,
	passing_info	*BrigdeInfo,
) error {
	cloner_path := filepath.Join(cloner_dir, cloner_name)
	log.Printf("Cloner path: %s", cloner_path)

	callback_path := filepath.Join(cloner_dir, "cage.callback")

	// init default value
	if passing_info == nil {
		passing_info = &BrigdeInfo{}
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

	//
	passing_info.Pipes = &BridgePipes{
		Stdout: stdout_pipe,
		Stderr: stderr_pipe,
		Result: result_pipe,
	}

	//
	content_string, err := func() (string, error) {
		var msgpack_bytes []byte
		enc := codec.NewEncoderBytes(&msgpack_bytes, &msgPackHandler)
		if err := enc.Encode(*passing_info); err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(msgpack_bytes), nil
	}()
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


	select {
	case stdout_err := <-stdout_c:
		_ = stdout_err
	case <-time.After(1 * time.Second):
	}

	select {
	case stderr_err := <-stderr_c:
		_ = stderr_err
	case <-time.After(1 * time.Second):
	}

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
			}

			cs <- err
			return;
		}

		log.Printf("= %d ==> %d", fd, size)
		log.Printf("= %d ==>\n%s\n<=====\n", fd, string(buffer))
	}

	cs <- nil
}



type Pipe struct {
	ReadFd, WriteFd		int
}

func makePipe() (*Pipe, error) {
	pipe := make([]int, 2)
	if err := syscall.Pipe(pipe); err != nil {
		return nil, err
	}

	return &Pipe{pipe[0], pipe[1]}, nil
}

func (p *Pipe) Close() {
	syscall.Close(p.ReadFd)
	syscall.Close(p.WriteFd)
}

type BridgePipes struct {
	Stdout, Stderr, Result	*Pipe
}



type JailedUserInfo struct {
	UserId		int
	GroupId		int
}

type BrigdeInfo struct {
	JailedUser			*JailedUserInfo
	Instruction			*ExecInstruction
	ChrootPath			string
	JailedUserHomePath	string
	IsReboot			bool
	Pipes				*BridgePipes
}


var msgPackHandler codec.MsgpackHandle

type sandboxCallback func(jailed_user *JailedUserInfo) error;
func sandboxBootstrap(
	callback sandboxCallback,
) error {
	expectRoot()

	user_name, uid, gid, err := CreateAnonUser()
	if err != nil {
		log.Printf("Couldn't create anon user")
		return err
	}
	defer func() {
		//
		killUserProcess(user_name, []string{"HUP", "KILL"})

		//
		const retry_times = 5
		succeeded := false
		for i:=0; i<retry_times; i++ {
			if err := DeleteUser(user_name); err != nil {
				log.Printf("Failed to delete user %s / %d times", user_name, i)
				killUserProcess(user_name, []string{"HUP", "KILL"})
			} else {
				succeeded = true
				break
			}
		}

		if !succeeded {
			// TODO: fix process...
			log.Printf("!! Failed to delete user %s for ALL", user_name)
		}
	}()

	//
	if callback != nil {
		err = callback(&JailedUserInfo{uid, gid})
	}

	return err
}


type ProcTarget struct {
	Id		int
	Version	string
}



type SourceData struct {
	Name			string
	Code			string
	IsCompressed	bool
}

type ExecDataset struct {
	StdinFilePath		*string
	CommandLine			string
	StructuredCommand	map[string]string
	CpuTimeLimit		uint64
	MemoryBytesLimit	uint64
}

type ExecInstruction struct {
	Profile		*ProcProfile
	Dataset		*ExecDataset
}


func (ctx *Context) build(
	base_name		string,
	sources			[]SourceData,
) error {
	err := sandboxBootstrap(
		func(jailed_user *JailedUserInfo) error {
			ctx.build2(base_name, sources, jailed_user)
			return nil
		})

	return err
}


func (ctx *Context) build2(
	base_name		string,
	sources			[]SourceData,
	jailed_user		*JailedUserInfo,
) {
	source_names := make([]string, len(sources))
	callbacks := make([]createTargetCallback, len(sources))

	for i, s := range sources {
		//
		source_names[i] = s.Name

		//
		callbacks[i] = func(*os.File) (error) {
			return nil
		}
	}

	//
	_, err := ctx.createMultipleTarget(base_name, jailed_user.GroupId, source_names, callbacks)
	if err != nil {
	}

	user_dir_path := ctx.makeUserDirName(base_name)
	user_home_path := ctx.jailedUserDir

	//
	compilation_info := &BrigdeInfo{
		JailedUser: jailed_user,
		Instruction: &ExecInstruction{
			&ProcProfile{
			},
			&ExecDataset{
				CpuTimeLimit: 10,
				MemoryBytesLimit: 1 * 1024 * 1024 * 1024,
			},
		},
		ChrootPath: user_dir_path,
		JailedUserHomePath: user_home_path,
		IsReboot: false,
	}
	compilation_info.invokeProcessCloner(filepath.Join(ctx.basePath, "bin"))


	//
	linking_info := &BrigdeInfo{
		JailedUser: jailed_user,
		Instruction: &ExecInstruction{
			&ProcProfile{
			},
			&ExecDataset{
			},
		},
		ChrootPath: user_dir_path,
		JailedUserHomePath: user_home_path,
		IsReboot: true,
	}
	linking_info.invokeProcessCloner(filepath.Join(ctx.basePath, "bin"))
}


func (bridge_info *BrigdeInfo) Hoge() error {
	if bridge_info.JailedUser == nil {
		return errors.New("Jailed User Info was NOT given")
	}

	if err := IntoJail(
		bridge_info.ChrootPath,
		bridge_info.JailedUserHomePath,
		bridge_info.IsReboot,
	); err != nil {
		return err
	}

	// Drop privilege
	if err := syscall.Setresgid(
		bridge_info.JailedUser.GroupId,
		bridge_info.JailedUser.GroupId,
		bridge_info.JailedUser.GroupId,
	); err != nil {
		return errors.New("Could NOT drop GROUP privilege")
	}

	if err := syscall.Setresuid(
		bridge_info.JailedUser.UserId,
		bridge_info.JailedUser.UserId,
		bridge_info.JailedUser.UserId,
	); err != nil {
		return errors.New("Could NOT drop USER privilege")
	}

	return nil
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

type ResourceLimit struct {
	CPU		uint64
	AS		uint64
	FSize	uint64
}

func (p *BridgePipes) execc(rl *ResourceLimit, command string, args []string, envs map[string]string) error {

	pid, err := fork()
	if err != nil {
		return err;
	}
	if pid == 0 {
		// child process
		defer os.Exit(-1)

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
			log.Print(path)
			if err := os.Setenv("PATH", path); err != nil {
				log.Fatal(err)
			}
		}

		//
		exec_path, err := exec.LookPath(command)
		if err != nil {
			log.Fatal(err)
		}

		//
		var env_list []string
		for k, v := range envs {
			env_list = append(env_list, k + "=" + v)
		}

		err = syscall.Exec(exec_path, append([]string{command}, args...), env_list);
		log.Fatal("unreachable : " + err.Error())
		return nil

	} else {
		// parent process

		//
		syscall.Close(p.Stdout.WriteFd)
		syscall.Close(p.Stderr.WriteFd)

		//
		process, err := os.FindProcess(pid)
		if err != nil {
			return err;
		}

		// parent process
		wait_pid_chan := make(chan *os.ProcessState)
		go func() {
			ps, _ := process.Wait()
			wait_pid_chan <- ps
		}()

		select {
		case ps := <-wait_pid_chan:
			usage, ok := ps.SysUsage().(*syscall.Rusage)
			if !ok {
				log.Fatal("akann")
			}
			fmt.Printf("%v", usage)

			// usage.Maxrss -> Amount of memory usage (KB)

		case <-time.After(time.Duration(rl.CPU * 2) * time.Second):
			// timeout(e.g. process uses sleep a lot)
		}


		return nil
	}
}



func (b *BrigdeInfo) Compile() {
	e := b.Instruction
	limit := &ResourceLimit{
		CPU: e.Dataset.CpuTimeLimit,		// CPU can be used only cpu_limit_time(sec)
		AS: e.Dataset.MemoryBytesLimit,		// Memory can be used only memory_limit_bytes
		FSize: 5 * 1024 * 1024,				// Process can writes a file only 5 MBytes
	}


	b.Pipes.execc(limit, "ls", []string{"-la", "/"}, map[string]string{"PATH": "/bin"})
}
