//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package main

import (
	"os"

	"yutopp/cage"
)

func main() {
	println("cage.callback booted")

	packed_torigoya_content := os.Getenv("packed_torigoya_content")

	if packed_torigoya_content == "" {
		panic("arguments are invalid")
	}

	//
	bm, err := torigoya.DecodeBridgeMessage(packed_torigoya_content)
	if err != nil {
		panic(err)
	}
	defer bm.Pipes.Close()

	// !!! ===================
	// Drop privilege
	// !!! ===================
	if err := bm.IntoJail(); err != nil {
		panic(err)
	}

	// execute given commands!
	if err := bm.Exec(); err != nil {
		panic(err)
	}
}
