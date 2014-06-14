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
	_ "strconv"
	_ "strings"
	_ "unicode"
	_ "errors"
)

func killUserProcess(
	user_name string,
	signals []string,
) error {
	for _, signal := range signals {
		err := exec.Command("pkill", "-", signal, "-u", user_name).Run()
		if err != nil {
			return err
		}
	}

	return nil
}
