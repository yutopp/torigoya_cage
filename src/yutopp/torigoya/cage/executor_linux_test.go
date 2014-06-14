//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import(
	"testing"
//	"syscall"
//	"os"
)

func TestIntoJail(t *testing.T) {
/*
	t.Logf("pid => %d\n", os.Getpid())

	// fork off the parent process
	ret, ret2, err := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	if err != 0 {
		t.Error("failed to fork")
		return
	}

	// failure
	if ret2 < 0 {
		t.Error("failed to fork")
		return
	}

	// if we got a good PID, then we call exit the parent process.
	if ret > 0 {
		t.Log("parent exit")
		return
	}

	t.Logf("pid => %d\n", os.Getpid())
	t.Error("test")


	//
	chroot_real_full_path := "/tmp/ticket/aaa"
	jailed_home := "home/torigoya"

	if err := intoJail(chroot_real_full_path, jailed_home); err != nil {
		t.Error(err.Error())
	}
*/
}
