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
	"errors"
	"os"
	"path/filepath"
)


func fileExists(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil
}

const implicitDefaultName = "*default*"

type TextContent struct {
	Name		string
	Data		[]byte
}


//
func (ctx *Context) createMultipleTargets(
	base_name			string,
	sources				[]*TextContent,
) (string, error) {
	return ctx.createMultipleTargetsWithDefaultName(
		base_name,
		sources,
		nil,
	)
}

func (ctx *Context) createMultipleTargetsWithDefaultName(
	base_name				string,
	sources					[]*TextContent,
	default_source_name		*string,
) (string, error) {
	log.Println(">> called createMultipleTargets")

	if len(sources) == 0 {
		return "", errors.New("inputs that length is 0 can not be accepted")
	}

	if !fileExists(ctx.userFilesBasePath) {
		return "", errors.New(fmt.Sprintf("directory %s is not existed", ctx.userFilesBasePath))
	}

	// ========================================
	// create the user directory
	user_home_dir_path := filepath.Join(ctx.userFilesBasePath, base_name)
	if err := ctx.craeteDir(user_home_dir_path); err != nil {
		return "", err
	}

	if err := ctx.craeteUserSources(
		user_home_dir_path,
		sources,
		default_source_name,
	); err != nil {
		return "", err
	}

	return user_home_dir_path, nil
}


func (ctx *Context) craeteDir(dir_name string) error {
	if fileExists(dir_name) {
		return errors.New(
			fmt.Sprintf(
				"directory %s is already existed",
				dir_name,
			),
		)
	}

	if err := os.Mkdir(dir_name, os.ModeDir); err != nil {
		return errors.New(
			fmt.Sprintf(
				"Couldn't create directory %s (err: %s)",
				dir_name,
				err,
			),
		)
	}

	// rwx/---/---
	if err := os.Chmod(dir_name, 0700); err != nil {
		return err
	}

	return nil
}

func (ctx *Context) craeteUserSources(
	user_home_dir_path	string,
	sources				[]*TextContent,
	default_source_name	*string,
) error {
	// ========================================
	//// make source file
	for _, source := range sources {
		if len(source.Name) == 0 {
			return errors.New("source_file_name must NOT be empty")
		}

		source_name := func() string {
			if default_source_name == nil {
				// if default_source_name is not specified, force to use source.Name
				return source.Name
			} else {
				if source.Name == implicitDefaultName {
					return *default_source_name
				} else {
					return source.Name
				}
			}
		}()

		source_full_path := filepath.Join(user_home_dir_path, source_name)
		f, err := os.OpenFile(source_full_path, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil { return err }
		defer f.Close()

		n, err := f.Write(source.Data)
		if err != nil { return err }
		if n != len(source.Data) { return errors.New("file length is different") }

		log.Printf("source -> %s\n", source_full_path)

		// r--/r--/---
		if err = os.Chmod(source_full_path, 0440); err != nil {
			return errors.New("failed to chmod")
		}
	}

	return nil
}
