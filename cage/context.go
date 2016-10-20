//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import (
	"errors"
	"fmt"
	"os"
)

type ContextOptions struct {
	BasePath          string
	UserFilesBasePath string
	SandboxExec       SandboxExecutor
}

type Context struct {
	basePath          string
	userFilesBasePath string
	sandboxExecutor   SandboxExecutor
}

func InitContext(opts *ContextOptions) (*Context, error) {
	// TODO: change to checking capability
	expectRoot()

	// create holder Directory, if not existed
	if !fileExists(opts.UserFilesBasePath) {
		err := os.Mkdir(opts.UserFilesBasePath, os.ModeDir|0700)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Couldn't create directory %s", opts.UserFilesBasePath))
		}
	}

	return &Context{
		basePath:          opts.BasePath,
		userFilesBasePath: opts.UserFilesBasePath,
		sandboxExecutor:   opts.SandboxExec,
	}, nil
}
