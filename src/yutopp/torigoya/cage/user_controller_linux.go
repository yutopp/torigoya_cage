//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import(
	"os/exec"
	"strconv"
	"strings"
	"unicode"
	"errors"

	"github.com/jmcvetta/randutil"
)


func CreateUser(user_name string) (int, int, error) {
	println("==> " + user_name)
	// create user
	user_craete_command := exec.Command("useradd", "--no-create-home", user_name)
	if err := user_craete_command.Run(); err != nil {
		return 0, 0, errors.New("Failed to useradd : " + err.Error())
	}

	// get uid/gid
	user_id_data, err := exec.Command("id", "--user", user_name).Output()
	if err != nil { return 0, 0, err }
	group_id_data, err := exec.Command("id", "--group", user_name).Output()
	if err != nil { return 0, 0, err }

	// convert ids from string to int
	user_id, err := strconv.Atoi(strings.TrimRightFunc(string(user_id_data), unicode.IsSpace))
	if err != nil { return 0, 0, err }
	group_id, err := strconv.Atoi(strings.TrimRightFunc(string(group_id_data), unicode.IsSpace))
	if err != nil { return 0, 0, err }

	if user_id == 0 || group_id == 0 {
		return 0, 0, errors.New("Invalid UserId or GroupId")
	}

	return user_id, group_id, nil
}


func DeleteUser(user_name string) error {
	user_delete_command := exec.Command("userdel", user_name)
	if err := user_delete_command.Run(); err != nil {
		return err
	}

	return nil
}

func CreateAnonUser() (string, int, int, error) {
	user_name, err := makeRandomUsername()
	if err != nil {
		return "", 0, 0, err
	}

	uid, gid, err := CreateUser(user_name)
	return user_name, uid, gid, err
}

// ([a-z_][a-z0-9_]{0,30})
func makeRandomUsername() (string, error) {
	piece, err := randutil.AlphaString(28)
	if err != nil {
		return "", err
	}

	return strings.ToLower("_" + piece), nil
}
