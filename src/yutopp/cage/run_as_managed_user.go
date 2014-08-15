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
	"log"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)


//
type JailedUserInfo struct {
	UserId		int
	GroupId		int
}


//
type runAsManagedUserCallback func(jailed_user *JailedUserInfo) error;

func runAsManagedUser(
	callback runAsManagedUserCallback,
) error {
	expectRoot()

	user_name, uid, gid, err := CreateAnonUser()
	if err != nil {
		log.Printf("Couldn't create anon user")
		return err
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Kill, syscall.SIGKILL, syscall.SIGINT, syscall.SIGHUP)
	defer func() {
		if err := recover(); err != nil {
            log.Printf("recoverd in runAsManagedUser: %v\n", err)
        }
		cleanupManagedUser(user_name)
		signal.Stop(sig)
	}()
	go func () {
		for _ = range sig {
			cleanupManagedUser(user_name)
			os.Exit(-1)		// prevent call defer
		}
	}()

	//
	if callback != nil {
		err = callback(&JailedUserInfo{ uid, gid })
	}

	return err
}

func cleanupManagedUser(user_name string) error {
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

		time.Sleep(10 * time.Millisecond)
	}

	if !succeeded {
		// TODO: fix process...
		log.Printf("!! Failed to delete user %s for ALL", user_name)
		return errors.New("Failed to delete user")
	}

	return nil
}
