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
	"time"
	"strconv"
)


func TestCreateUser(t *testing.T) {
	test_user := "torigoya_hoge" + strconv.Itoa(time.Now().Nanosecond())

	t.Logf("Create user = %s", test_user)
	uid, gid, err := CreateUser(test_user)
	if err != nil {
		t.Errorf("ababa" + err.Error())
	}
	if uid == 0 {
		t.Errorf("invalid uid %d", uid)
	}
	if gid == 0 {
		t.Errorf("invalid gid %d", uid)
	}
	t.Logf("Created user = %s, uid = %d, gid = %d", test_user, uid, gid)

	t.Logf("Delete user = %s", test_user)
	if err := DeleteUser(test_user); err != nil {
		t.Errorf("ababa" + err.Error())
	}
}
