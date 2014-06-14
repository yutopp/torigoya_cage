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

	"yutopp/torigoya/cage"

	"encoding/base64"
	"github.com/ugorji/go/codec"
)

var msgPackHandler codec.MsgpackHandle

func main() {
	println("cage.callback booted")

	packed_torigoya_content := os.Getenv("packed_torigoya_content")

	if packed_torigoya_content == "" {
		panic("arguments are invalid")
	}

	decoded_bytes, err := base64.StdEncoding.DecodeString(packed_torigoya_content)
	if err != nil {
		panic(err)
	}

	var bridge_info torigoya.BrigdeInfo
	dec := codec.NewDecoderBytes(decoded_bytes, &msgPackHandler)
	err = dec.Decode(&bridge_info)
	if err != nil {
		panic(err)
	}

	// !!! ===================
	// Drop privilege
	// !!! ===================
	if err := bridge_info.Hoge(); err != nil {
		panic(err)
	}

	bridge_info.Compile()
}
