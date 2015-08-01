//
// Copyright yutopp 2015 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

// +build linux

package torigoya

import(
	"log"
	"errors"
    "os/exec"
)

func (ctx *Context) UpdatePackages() error {
	if ctx.packageUpdater == nil {
		return errors.New("Package Updater was not registerd")
	}

	err := ctx.packageUpdater.Update()
	if err != nil {
		return err
	}

	// TODO: fix it
    log.Printf("= %s ============================", ctx.packageUpdater.GetInstallPrefix())
	out, err := exec.Command("/bin/ls", "-la", ctx.packageUpdater.GetInstallPrefix()).Output()
	if err != nil {
		log.Printf("error:: %s", err.Error())
	} else {
		log.Printf("package update passed:: %s", out)
	}
	log.Printf("==================================================\n")

    return err
}
