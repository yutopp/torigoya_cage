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
	"os"
	"syscall"
	"errors"
	"fmt"
	"log"
)

// #define _BSD_SOURCE
// #include <sys/types.h>
// int devno(int major, int minor)
// {
//     return makedev( major, minor );
// }
import "C"


//
func (bm *BridgeMessage) IntoJail() error {
	if bm.JailedUser == nil {
		return errors.New("Jailed User Info was NOT given")
	}

	//
	if err := buildChrootEnv(
		bm.ChrootPath,
		bm.JailedUserHomePath,
		bm.IsReboot,
	); err != nil {
		return err
	}

	// Drop privilege(group)
	if err := syscall.Setresgid(
		bm.JailedUser.GroupId,
		bm.JailedUser.GroupId,
		bm.JailedUser.GroupId,
	); err != nil {
		return errors.New("Could NOT drop GROUP privilege")
	}

	// Drop privilege(user)
	if err := syscall.Setresuid(
		bm.JailedUser.UserId,
		bm.JailedUser.UserId,
		bm.JailedUser.UserId,
	); err != nil {
		return errors.New("Could NOT drop USER privilege")
	}

	return nil
}


// mount system's
// http://linuxjm.sourceforge.jp/html/LDP_man-pages/man2/mount.2.html
var readOnlyMounts = []string {
	"/etc",
	"/include",
	"/lib",
	"/lib32",
	"/lib64",
	"/bin",
	"/usr/include",
	"/usr/lib",
	"/usr/lib32",
	"/usr/lib64",
	"/usr/bin",
	"/usr/local/torigoya",
}


func buildChrootEnv(
	chroot_root_full_path	string,
	jail_home				string,
	only_chroot				bool,
) (err error) {
    log.Printf("buildChrootEnv::chroot_root_full_path: %s\n", chroot_root_full_path);
    log.Printf("buildChrootEnv::jail_home: %s\n", jail_home);

	expectRoot()

	//
	if err := os.Chdir(chroot_root_full_path); err != nil {
		return errors.New(fmt.Sprintf("failed to chdir -> %s (%s)", chroot_root_full_path, err))
	}

	//
	if !only_chroot {
		// mount system's
		for _, host_mount_name := range readOnlyMounts {
			if !fileExists(host_mount_name) {
				log.Printf("system dir %s is not existed on host machine\n", host_mount_name)
				continue
			}

			local_mount_name := "." + host_mount_name
			err := os.MkdirAll(local_mount_name, 0555)
			if err != nil {
				return errors.New(fmt.Sprintf("failed to mkdir -> %s (%s)", local_mount_name, err))
			}

			err = syscall.Mount(
				host_mount_name,
				local_mount_name,
				"",
				syscall.MS_BIND | syscall.MS_RDONLY | syscall.MS_NOSUID | syscall.MS_NODEV,
				"",
			)

			println(host_mount_name)
		}


		// mount procfs
		if err := os.MkdirAll("proc", 0755); err != nil {
			return errors.New(fmt.Sprintf("failed to mkdir proc (%s)", err))
		}
		if err := syscall.Mount(
			"/proc",
			"./proc",
			"proc",
			syscall.MS_BIND | syscall.MS_RDONLY | syscall.MS_NOSUID | syscall.MS_NODEV,
			"",
		); err != nil {
			return errors.New(fmt.Sprintf("failed to mount /proc -> ./proc (%s)", err))
		}


		// mount /tmp
		if err := os.MkdirAll("tmp", 0777); err != nil {
			return errors.New(fmt.Sprintf("failed to mkdir tmp (%s)", err))
		}
		if err := syscall.Mount(
			"",
			"./tmp",
			"tmpfs",
			syscall.MS_NOEXEC | syscall.MS_NODEV,
			"",
		); err != nil {
			return errors.New(fmt.Sprintf("failed to mount /tmp -> ./tmp (%s)", err))
		}

		// create /dev
		if err := os.MkdirAll("dev", 0555); err != nil {
			return errors.New(fmt.Sprintf("failed to mkdir dev (%s)", err))
		}

		if err := syscall.Mknod("dev/null", syscall.S_IFCHR|0666, int(C.devno(1, 3))); err != nil {
			return errors.New(fmt.Sprintf("failed to mknod dev/null (%s)", err))
		}
		if err := syscall.Mknod("dev/zero", syscall.S_IFCHR|0666, int(C.devno(1, 5))); err != nil {
			return errors.New(fmt.Sprintf("failed to mknod dev/zero (%s)", err))
		}
		if err := syscall.Mknod("dev/full", syscall.S_IFCHR|0666, int(C.devno(1, 7))); err != nil {
			return errors.New(fmt.Sprintf("failed to mknod dev/full (%s)", err))
		}
		if err := syscall.Mknod("dev/random", syscall.S_IFCHR|0644, int(C.devno(1, 8))); err != nil {
			return errors.New(fmt.Sprintf("failed to mknod dev/random (%s)", err))
		}
		if err := syscall.Mknod("dev/urandom", syscall.S_IFCHR|0644, int(C.devno(1, 9))); err != nil {
			return errors.New(fmt.Sprintf("failed to mknod dev/urandom (%s)", err))
		}
	}


	// DO chroot !!
	if err := syscall.Chroot(chroot_root_full_path); err != nil {
		return errors.New(fmt.Sprintf("failed to chroot (%s)", err))
	}


	if !only_chroot {
		//
		if err := os.Symlink("/proc/self/fd/0", "/dev/stdin"); err != nil {
			return errors.New(fmt.Sprintf("failed to symlink (%s)", err))
		}

		if err := os.Symlink("/proc/self/fd/1", "/dev/stdout"); err != nil {
			return errors.New(fmt.Sprintf("failed to symlink (%s)", err))
		}
		if err := os.Symlink("/proc/self/fd/2", "/dev/stderr"); err != nil {
			return errors.New(fmt.Sprintf("failed to symlink (%s)", err))
		}
	}


	// change current directory to TARGET path
	if err := os.Chdir(jail_home); err != nil {
		return errors.New(fmt.Sprintf("failed to chdir -> %s (%s)", jail_home, err))
	}

	log.Printf("<- buildChrootEnv");

	return nil
}
