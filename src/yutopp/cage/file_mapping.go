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
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)


func fileExists(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil
}


func (ctx *Context) makeUserDirName(base_name string) string {
	return filepath.Join(ctx.sandboxDir, base_name)
}


type TextContent struct {
	Name		string
	Data		[]byte
}

func (ctx *Context) createTarget(
	base_name			string,
	managed_group_id	int,
	source				*TextContent,
) (string, error) {
	source_full_paths, err := ctx.createMultipleTargets(
		base_name,
		managed_group_id,
		[]*TextContent{ source },
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
func (ctx *Context) createMultipleTargets(
	base_name			string,
	managed_group_id	int,
	sources				[]*TextContent,
) (source_full_paths []string, err error) {
	return ctx.createMultipleTargetsWithDefaultName(
		base_name,
		managed_group_id,
		sources,
		nil,
	)
}

//
func (ctx *Context) createMultipleTargetsWithDefaultName(
	base_name			string,
	managed_group_id	int,
	sources				[]*TextContent,
	default_name		*string,
) (source_full_paths []string, err error) {
	log.Println(">> called createMultipleTargets")

	//
	if len(sources) == 0 {
		return nil, errors.New("inputs that length is 0 can not be accepted")
	}

	//
    expectRoot()

	log.Printf("Euid -> %d\n", os.Geteuid())

	// In posix, Uid only contains numbers
	host_user_id, _ := strconv.Atoi(ctx.hostUser.Uid)
	log.Printf("Host uid: %s\n", ctx.hostUser.Uid)

	//
	if !fileExists(ctx.sandboxDir) {
		return nil, errors.New(fmt.Sprintf("directory %s is not existed", ctx.sandboxDir))
	}

	// ========================================
	//// create user directory

	//
	user_dir_path := ctx.makeUserDirName(base_name)

	//
	if fileExists(user_dir_path) {
		log.Printf("user directory %s is already existed, so remove them\n", user_dir_path)
//		if err := umountJail(user_dir_path); err != nil {
//			return nil, errors.New(fmt.Sprintf("Couldn't unmount directory %s (%s)", user_dir_path, err))
//		}
		if err := os.RemoveAll(user_dir_path); err != nil {
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
	// host_user_id:managed_group_id // r-x/r-x/---
	if err := guardPath(user_home_base_path, host_user_id, managed_group_id, 0550); err != nil {
		return nil, err
	}

	// create /home/torigoya
	user_home_path := filepath.Join(user_dir_path, ctx.jailedUserDir)
	if err := os.Mkdir(user_home_path, os.ModeDir); err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't create directory %s (%s)", user_home_path, err))
	}
	// host_user_id:managed_group_id // rwx/rwx/---
	// NOTE: add "write" permission to group to output executable files
	if err := guardPath(user_home_path, host_user_id, managed_group_id, 0770); err != nil {
		return nil, err
	}

	// ========================================
	//// make source file
	source_full_paths = make([]string, len(sources))
	for index, source := range sources {
		if len(source.Name) == 0 {
			return nil, errors.New("source_file_name must NOT be empty")
		}

		source_name := func() string {
			if default_name == nil {
				return source.Name
			} else {
				if source.Name == "*default*" {
					return *default_name
				} else {
					return source.Name
				}
			}
		}()

		source_full_path := filepath.Join(user_home_path, source_name)
		f, err := os.OpenFile(source_full_path, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return nil, err
		}
		defer func() {
			f.Close()
			log.Printf("-> %s\n", source_full_path)
			// host_user_id:managed_group_id // r--/r--/---
			err = guardPath(source_full_path, host_user_id, managed_group_id, 0440)
		}()

		//
		_, err = f.Write(source.Data)

		//
		source_full_paths[index] = source_full_path
	}

	log.Printf("==================================================\n")
	out, err := exec.Command("/bin/ls", "-laR", user_home_path).Output()
	if err != nil {
		log.Printf("error:: %s\n", err.Error())
	} else {
		log.Printf("passed:: %s\n", out)
	}

	return source_full_paths, err
}


//
type reassignTargetCallback func(string) (*string, error)

func (ctx *Context) reassignTarget(
	base_name				string,
	managed_user_id			int,
	managed_group_id		int,
	callback				reassignTargetCallback,
) (user_dir_path string, input_path *string, err error) {
	log.Println("called SekiseiRunnerNodeServer::reassign_target")

    expectRoot()

	// In posix, Uid only contains numbers
	host_user_id, _ := strconv.Atoi(ctx.hostUser.Uid)
	log.Printf("host uid: %s\n", ctx.hostUser.Uid)

	if err := ctx.cleanupMountedFiles(base_name); err != nil {
		return "", nil, err
	}

	//
	user_dir_path = ctx.makeUserDirName(base_name)

	//
	if err := guardPath(user_dir_path, host_user_id, managed_group_id, 0550); err != nil {
		return "", nil, err
	}

	// chmod /home // host_user_id:managed_group_id // r-x/r-x/---
	user_home_base_path := filepath.Join(user_dir_path, ctx.homeDir)
	if err := guardPath(user_home_base_path, host_user_id, managed_group_id, 0550); err != nil {
		return "", nil, err
	}

	// chmod /home/torigoya
	user_home_path := filepath.Join(user_dir_path, ctx.jailedUserDir)
	// host_user_id:managed_group_id // rwx/---/---
	if err := guardPath(user_home_path, host_user_id, managed_group_id, 0700); err != nil {
		return "", nil, err
	}

	// call user block
	if callback != nil {
		input_path, err = callback(user_dir_path)
		if err != nil {
			return "", nil, err
		}
	}

	// host_user_id:managed_group_id // rwx/rwx/---
	if err := guardPath(user_home_path, host_user_id, managed_group_id, 0770); err != nil {
		return "", nil, err
	}

	//
	err = filepath.Walk(user_home_path, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if !info.IsDir() {
			if info.Mode() & 0100 != 0 {
				// specialize if has permission --x/---/---
				if err := os.Chown(path, managed_user_id, managed_group_id); err != nil {
					return errors.New(fmt.Sprintf("Couldn't chown %s, %s", path, err.Error()))
				}

				log.Printf("reassgin::chown %s -> %d : %d \n", path, managed_user_id, managed_group_id)

			} else {
				//
				if err := os.Chown(path, host_user_id, managed_group_id); err != nil {
					return errors.New(fmt.Sprintf("Couldn't chown %s, %s", path, err.Error()))
				}

				log.Printf("reassgin::chown %s -> %d : %d \n", path, host_user_id, managed_group_id)
			}
		}
		return err
	})

	log.Printf("==================================================\n")
	out, err := exec.Command("/bin/ls", "-laR", user_home_path).Output()
	if err != nil {
		log.Printf("error:: %s\n", err.Error())
	} else {
		log.Printf("passed:: %s\n", out)
	}

	return user_dir_path, input_path, err
}


func (ctx *Context) cleanupMountedFiles(
	base_name				string,
) error {
	log.Println("called file_mapping::cleanupMountedFiles")

    expectRoot()

	//
	user_dir_path := ctx.makeUserDirName(base_name)

	// delete directories exclude HOME
	// TODO: DO NOT DELETE directories that mount HOST directories
	dirs, err := filepath.Glob(filepath.Join(user_dir_path, "/*"))
	if err != nil { return err }

//	if err := umountJail(user_dir_path); err != nil {
//		return errors.New(fmt.Sprintf("Couldn't unmount directory %s (%s)", user_dir_path, err))
//	}

	for _, dir := range dirs {
		rel_dir, err := filepath.Rel(user_dir_path, dir)
		if err != nil { return err }

		if rel_dir != ctx.homeDir {
			if err := os.RemoveAll(dir); err != nil {
				return errors.New(fmt.Sprintf("Couldn't remove directory %s (%s)", dir, err))
			}
		}
	}

	return nil
}


func (ctx *Context) createInput(
	base_dir_path		string,
	managed_group_id	int,
	stdin				*TextContent,
) (stdin_full_path string, err error) {
	log.Println("called SekiseiRunnerNodeServer::createInput")

    expectRoot()

	// In posix, Uid only contains numbers
	host_user_id, _ := strconv.Atoi(ctx.hostUser.Uid)
	log.Printf("host uid: %s\n", ctx.hostUser.Uid)

	//
	const inputs_dir_name = "stdin"
	inputs_dir_path := filepath.Join(base_dir_path, ctx.jailedUserDir, inputs_dir_name)

	//
	if !fileExists(inputs_dir_path) {
		err := os.Mkdir(inputs_dir_path, os.ModeDir)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Couldn't create directory %s", inputs_dir_path))
		}
	}
	// host_user_id:managed_group_id // rwx/---/---
	if err := guardPath(inputs_dir_path, host_user_id, managed_group_id, 0700); err != nil {
		return "", err
	}

	//
	stdin_full_path = filepath.Join(inputs_dir_path, stdin.Name)
	f, err := os.OpenFile(stdin_full_path, os.O_WRONLY|os.O_CREATE, 0440)	// r--/r--/---
	if err != nil {
		return "", err
	}
	defer func() {
 		f.Close()
		// host_user_id:managed_group_id // r--/r--/---
		err = guardPath(stdin_full_path, host_user_id, managed_group_id, 0440)
	}()
	if _, err := f.Write(stdin.Data); err != nil {
		return "", err
	}

	// change input DIR permission
	// host_user_id:managed_group_id // r-x/r-x/---
	if err := guardPath(inputs_dir_path, host_user_id, managed_group_id, 0550); err != nil {
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
