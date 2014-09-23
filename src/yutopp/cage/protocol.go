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

type MessageKind	uint8
const (
	MessageKindIndexBegin				= MessageKind(0)

	// Sent from client
	MessageKindAcceptRequest			= MessageKind(0)
	MessageKindTicketRequest			= MessageKind(1)
	MessageKindUpdateRepositoryRequest	= MessageKind(2)
	MessageKindReloadProcTableRequest	= MessageKind(3)
	MessageKindUpdateProcTableRequest	= MessageKind(4)
	MessageKindGetProcTableRequest		= MessageKind(5)

	// Sent from server
	MessageKindAccept					= MessageKind(6)
	MessageKindOutputs					= MessageKind(7)
	MessageKindResult					= MessageKind(8)
	MessageKindSystemError				= MessageKind(9)
	MessageKindExit						= MessageKind(10)

	MessageKindSystemResult				= MessageKind(11)
	MessageKindProcTable				= MessageKind(12)

	//
	MessageKindIndexEnd					= MessageKind(12)
	MessageKindInvalid					= MessageKind(0xff)
)


func (k MessageKind) String() string {
	switch k {
	case MessageKindAcceptRequest:
		return "MessageKindAcceptRequest"
	case MessageKindTicketRequest:
		return "MessageKindTicketRequest"
	case MessageKindUpdateRepositoryRequest:
		return "MessageKindUpdateRepositoryRequest"
	case MessageKindReloadProcTableRequest:
		return "MessageKindReloadProcTableRequest"
	case MessageKindUpdateProcTableRequest:
		return "MessageKindUpdateProcTableRequest"
	case MessageKindGetProcTableRequest:
		return "MessageKindGetProcTableRequest"
	default:
		return fmt.Sprintf("%d", k)
	}
}

// torigoya protocol
// header 5 bytes
// [header(1bytes)|length of data(uint, little endian 4bytes)|data(msgpacked)]
const HeaderLength = 5

//
type ProtocolDataType map[string]string

//
func EncodeToTorigoyaProtocol(kind MessageKind, data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	// write kind(1Bytes)
	if kind < MessageKindIndexBegin || kind > MessageKindIndexEnd {
		return nil, errors.New(fmt.Sprintf("Failed to write data / invalid header %d", kind))
	}
	if err := binary.Write(buf, binary.LittleEndian, kind); err != nil { return nil, err }

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

func (ph *ProtocolHandler) read(reader io.Reader) (MessageKind, interface{}, error) {
	// read protocol

	// read header
	n, err := reader.Read(ph.header_buffer[:])
	if err != nil { return MessageKindInvalid, nil, err }
	if n < HeaderLength {
		return MessageKindInvalid, nil, errors.New("invalid header length")
	}
	log.Printf("read length:%d / bal: %v\n", n, ph.header_buffer[:n])

	// kind
	kind := ph.header_buffer[0]

	// length of data
	var length uint32
	if err := binary.Read(bytes.NewReader(ph.header_buffer[1:]), binary.LittleEndian, &length); err != nil {
		return MessageKindInvalid, nil, err
	}
	log.Printf("length of data: %d\n", length)

	// source code limit: 256KB
	if length > 256 * 1024 {
		return MessageKindInvalid, nil, errors.New("SourceCode length limitation")
	}

	//
	if uint32(len(ph.buffer)) < length {
		ph.buffer = make([]byte, length)
	}
	n, err = io.ReadFull(reader, ph.buffer[0:length])
	if err != nil {
		return MessageKindInvalid, nil, err
	}
	if uint32(n) != length {
		return MessageKindInvalid, nil, errors.New(fmt.Sprintf("%d", n))
	}

	//
	log.Printf("read:: kind: %d / length: %d, /value: %v\n", kind, length, ph.buffer[0:length])
	var data interface{}
	dec := codec.NewDecoderBytes(ph.buffer[0:length], &msgPackHandler)
	if err := dec.Decode(&data); err != nil {
		return MessageKindInvalid, nil, err
	}

	return MessageKind(kind), data, nil
}

func (ph *ProtocolHandler) write(writer io.Writer, header MessageKind, data interface{}) error {
	buf, err := EncodeToTorigoyaProtocol(header, data)
	if err != nil {
		return err
	}

	//log.Printf("write::value: %v\n", buf)

	n, err := writer.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return errors.New("couldn't send all bytes")
	}

	return nil
}


//
func (ph *ProtocolHandler) writeAccept(
	writer io.Writer,
) error {
	return ph.write(writer, MessageKindAccept, nil)
}

func (ph *ProtocolHandler) writeOutputResult(
	writer io.Writer,
	r *StreamOutputResult,
) error {
	return ph.write(writer, MessageKindOutputs, r.ToTuple())
}

//
func (ph *ProtocolHandler) writeExecutedResult(
	writer io.Writer,
	r *StreamExecutedResult,
) error {
	return ph.write(writer, MessageKindResult, r.ToTuple())
}

//
func (ph *ProtocolHandler) writeSystemError(
	writer io.Writer,
	message string,
) error {
	return ph.write(writer, MessageKindSystemError, message)
}

//
func (ph *ProtocolHandler) writeExit(
	writer io.Writer,
) error {
	return ph.write(writer, MessageKindExit, "")
}


//
func (ph *ProtocolHandler) writeSystemResult(
	writer io.Writer,
	status int,
) error {
	return ph.write(writer, MessageKindSystemResult, status)
}

//
func (ph *ProtocolHandler) writeProcTable(
	writer				io.Writer,
	proc_config_table	*ProcConfigTable,
) error {
	return ph.write(writer, MessageKindProcTable, proc_config_table)
}
