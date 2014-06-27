//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import (
	"testing"
	"os"
	"fmt"
	"net"

	"github.com/ugorji/go/codec"
)


func TestProtocol(t *testing.T) {
	data := ProtocolDataType{ "abcde": "buffer" }

	result, err := EncodeToTorigoyaProtocol(HeaderRequest, data)

	_ = result
	_ = err

	if len(result) == 0 {
		t.Fatalf("length should be 0 (but %d)", len(result))
	}
/*
	if !bytes.Equal() {
	}
*/
	t.Fatalf("||| %v", result)
}


func TestProtocolServer(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	ctx, err := InitContext(gopath)
	if err != nil {
		t.Fatalf(err.Error())
	}

	e := make(chan error)
	go RunServer(":12321", ctx, e)
	if err := <- e; err != nil {
		t.Fatal(err)
	}

	conn, err := net.Dial("tcp", ":12321")
	if err != nil {
		t.Fatal(err)
		// handle error
	}

	//
	var handler ProtocolHandler

	// request
	if err := handler.WriteRequest(conn, "aba"); err != nil {
		t.Fatalf("server recv: %v\n", err)
	}


	//
	for {
		kind, data, err := handler.Read(conn)
		if err != nil {
			t.Fatalf("client error: %v\n", err)
			break
		}

		fmt.Printf("client recv: %d / %v\n", kind, data)
	}
}


func TestProtocolReadTicketFromPackedData(t *testing.T) {
	// packed data
	buffer := []byte{ 150, 163, 97, 97, 97, 0, 165, 48, 46, 48, 46, 48, 145, 147, 168, 112, 114, 111, 103, 46, 99, 112, 112, 163, 97, 97, 97, 194, 146, 148, 160, 145, 146, 161, 97, 161, 98, 10, 206, 32, 0, 0, 0, 148, 160, 145, 146, 161, 97, 161, 98, 10, 206, 32, 0, 0, 0, 145, 145, 146, 192, 148, 160, 145, 146, 161, 97, 161, 98, 10, 206, 32, 0, 0, 0 }

	// decode
	var data interface{}
	dec := codec.NewDecoderBytes(buffer, &msgPackHandler)
	if err := dec.Decode(&data); err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Printf("data %V\n", data)

	// construct
	ticker, err := MakeTicketFromTuple(data)
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Printf("data %V\n", ticker)

	//
}
