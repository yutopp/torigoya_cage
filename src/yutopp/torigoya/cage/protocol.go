//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import (
	"fmt"
	"io"
	"bytes"
    "encoding/binary"
	"errors"
	"log"

	"github.com/ugorji/go/codec"
)


const (
	HeaderIndexBegin	= 0

	HeaderRequest		= 0
	HeaderOutputs		= 1
	HeaderResult		= 2
	HeaderSystemError	= 3
	HeaderExit			= 4

	HeaderIndexEnd		= 4

	HeaderInvalid		= 0xff
)

var HeaderString = []string{ "REQ", "OUT", "RES" , "ERR", "END" }

// torigoya protocol
// header 5 bytes
// [header(1bytes)|length of data(uint, little endian 4bytes)|data(msgpacked)]
const HeaderLength = 5

//
type ProtocolDataType map[string]string

//
func EncodeToTorigoyaProtocol(header int8, data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	// make header(1Bytes)
	if header < HeaderIndexBegin || header > HeaderIndexEnd {
		return nil, errors.New(fmt.Sprintf("Failed to write data / invalid header %d", header))
	}
	if err := binary.Write(buf, binary.LittleEndian, header); err != nil { return nil, err }

	// (encode data)
	var msgpack_bytes []byte
	enc := codec.NewEncoderBytes(&msgpack_bytes, &msgPackHandler)
	if err := enc.Encode(&data); err != nil {
		return nil, err
	}

	// length
	length := uint32(len(msgpack_bytes))
	err := binary.Write(buf, binary.LittleEndian, length)
	if err != nil { return nil, err }

	// data
	{
		length, err := buf.Write(msgpack_bytes)
		if err != nil { return nil, err }
		if length != length { return nil, errors.New("Failed to write data") }
	}

	return buf.Bytes(), nil
}




type ProtocolHandler struct {
	header_buffer	[HeaderLength]byte
	buffer			[]byte
}

func (ph *ProtocolHandler) read(reader io.Reader) (uint8, interface{}, error) {
	// read protocol
	n, err := reader.Read(ph.header_buffer[:])
	if err != nil { return HeaderInvalid, nil, err }
	if n < HeaderLength {
		return HeaderInvalid, nil, errors.New("")
	}
	log.Printf("read length:%d / bal: %v\n", n, ph.header_buffer[:n])

	// kind
	kind := ph.header_buffer[0]

	// length of data
	var length uint32
	if err := binary.Read(bytes.NewReader(ph.header_buffer[1:]), binary.LittleEndian, &length); err != nil {
		return HeaderInvalid, nil, err
	}

	//
	if uint32(len(ph.buffer)) < length {
		ph.buffer = make([]byte, length)
	}
	n, err = reader.Read(ph.buffer[:])
	if err != nil {
		return HeaderInvalid, nil, err
	}
	if uint32(n) != length {
		return HeaderInvalid, nil, errors.New("")
	}

	//
	log.Printf("read:: kind: %d / length: %d, /value: %v\n", kind, length, ph.buffer)
	var data interface{}
	dec := codec.NewDecoderBytes(ph.buffer[:], &msgPackHandler)
	if err := dec.Decode(&data); err != nil {
		return HeaderInvalid, nil, err
	}

	return kind, data, nil
}

func (ph *ProtocolHandler) write(writer io.Writer, header int8, data interface{}) error {
	buf, err := EncodeToTorigoyaProtocol(header, data)
	if err != nil {
		return err
	}

	log.Printf("write::value: %v\n", buf)

	n, err := writer.Write(buf)
	if err != nil {
		return err
	}

	if n != len(buf) {
		return errors.New("")
	}

	return nil
}

func (ph *ProtocolHandler) WriteRequest(writer io.Writer, message *Ticket) error {
	return ph.write(writer, HeaderRequest, message)
}


func (ph *ProtocolHandler) WriteError(writer io.Writer, message string) error {
	return ph.write(writer, HeaderResult, message)
}

func (ph *ProtocolHandler) WriteExit(writer io.Writer) error {
	return ph.write(writer, HeaderExit, "")
}
