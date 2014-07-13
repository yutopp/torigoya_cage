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
	"errors"
	"strconv"
	"os"
	"os/user"
	"path/filepath"
)

type PackageUpdater interface {
	Update() error
}

type Context struct {
	basePath			string

	hostUser			*user.User

	sandboxDir			string
	homeDir				string
	jailedUserDir		string

	procConfPath		string
	procConfTable		ProcConfigTable

	procSrcZipAddress	string
	packageUpdater		PackageUpdater
}


func InitContext(
	base_path				string,
	host_user_name			string,
	proc_config_path		string,
	proc_src_zip_address	string,
	package_updater			PackageUpdater,
) (*Context, error) {
	// TODO: change to checking capability
	expectRoot()

	//
	sandbox_dir := "/tmp/sandbox"

	//
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
			return nil, errors.New(fmt.Sprintf("Couldn't create directory %s", sandbox_dir))
		}

		if err := filepath.Walk(sandbox_dir, func(path string, info os.FileInfo, err error) error {
			if err != nil { return err }
			// r-x/---/---
			err = guardPath(path, host_user_id, host_user_id, 0500)
			return err
		}); err != nil {
			return nil, errors.New(fmt.Sprintf("Couldn't create directory %s", sandbox_dir))
		}
	}

	// LoadProcConfigTable
	proc_conf_table, err := LoadProcConfigs(proc_config_path)
	if err != nil {
		return nil, err
	}

	//
	return &Context{
		basePath:			base_path,
		hostUser:			host_user,
		sandboxDir:			sandbox_dir,
		homeDir:			"home",
		jailedUserDir:		"home/torigoya",
		procConfPath:		proc_config_path,
		procConfTable:		proc_conf_table,
		procSrcZipAddress:	proc_src_zip_address,
		packageUpdater:		package_updater,
	}, nil
}


func (ctx *Context) UpdatePackages() error {
	if ctx.packageUpdater == nil {
		return errors.New("Package Updater was not registerd")
	}

	return ctx.packageUpdater.Update()
}


func (ctx *Context) ReloadProcTable() error {
	// RELOAD LoadProcConfigTable
	proc_conf_table, err := LoadProcConfigs(ctx.procConfPath)
	if err != nil {
		return err
	}

	// rewrite
	ctx.procConfTable = proc_conf_table

	return nil
}

func (ctx *Context) UpdateProcTable() error {
	return ctx.procConfTable.UpdateFromWeb(ctx.procSrcZipAddress)
}
