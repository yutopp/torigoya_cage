//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

// +build linux

package torigoya

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

const implicitDefaultName = "*default*"

type TextContent struct {
	Name string
	Data []byte
}

//
func (ctx *Context) createMultipleTargets(
	baseName string,
	sources []*TextContent,
) (string, error) {
	return ctx.createMultipleTargetsWithDefaultName(
		baseName,
		sources,
		nil,
	)
}

func (ctx *Context) createMultipleTargetsWithDefaultName(
	baseName string,
	sources []*TextContent,
	default_source_name *string,
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
	userHomeDirPath := filepath.Join(ctx.userFilesBasePath, baseName)
	if err := ctx.createDir(userHomeDirPath); err != nil {
		return "", err
	}

	if err := ctx.createUserSources(
		userHomeDirPath,
		sources,
		default_source_name,
	); err != nil {
		return "", err
	}

	return userHomeDirPath, nil
}

func (ctx *Context) createDir(dirName string) error {
	if fileExists(dirName) {
		return fmt.Errorf(
			"directory %s is already existed",
			dirName,
		)
	}

	if err := os.Mkdir(dirName, os.ModeDir); err != nil {
		return errors.New(
			fmt.Sprintf(
				"Couldn't create directory %s (err: %s)",
				dirName,
				err,
			),
		)
	}

	// rwx/---/---
	if err := os.Chmod(dirName, 0700); err != nil {
		return err
	}

	return nil
}

func (ctx *Context) createUserSources(
	userHomeDirPath string,
	sources []*TextContent,
	default_source_name *string,
) error {
	for _, source := range sources {
		if err := ctx.createUserSource(userHomeDirPath, source); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *Context) createUserSource(
	userHomeDirPath string,
	source *TextContent,
) error {
	if len(source.Name) == 0 {
		return errors.New("source_file_name must NOT be empty")
	}

	sourceFullPath := filepath.Join(userHomeDirPath, filepath.Clean(source.Name))
	f, err := os.OpenFile(sourceFullPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := f.Write(source.Data)
	if err != nil {
		return err
	}
	if n != len(source.Data) {
		return errors.New("file length is different")
	}

	log.Printf("source -> %s\n", sourceFullPath)

	// r--/r--/---
	if err = os.Chmod(sourceFullPath, 0440); err != nil {
		return errors.New("failed to chmod")
	}

	return nil
}
