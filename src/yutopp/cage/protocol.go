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
	"bytes"
    "encoding/binary"
	"errors"
	"io"

	"github.com/ugorji/go/codec"
)

type MessageKind	uint8
const (
	MessageKindIndexBegin				= MessageKind(0)

	// Sent from client
	MessageKindTicketRequest			= MessageKind(1)
	MessageKindUpdateRepositoryRequest	= MessageKind(2)

	// Sent from server
	MessageKindOutputs					= MessageKind(7)
	MessageKindResult					= MessageKind(8)
	MessageKindSystemError				= MessageKind(9)
	MessageKindExit						= MessageKind(10)

	MessageKindSystemResult				= MessageKind(11)

	//
	MessageKindIndexEnd					= MessageKind(11)
	MessageKindInvalid					= MessageKind(0xff)
)


func (k MessageKind) String() string {
	switch k {
	case MessageKindTicketRequest:
		return "MessageKindTicketRequest"
	case MessageKindUpdateRepositoryRequest:
		return "MessageKindUpdateRepositoryRequest"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

var TorigoyaProtocolSignature = [2]byte{0x54, 0x47}	// TP
// Signature		[2]byte		//
// MessageKind		byte		//
// Version			[4]byte		// uint32, little endian
// Length			[4]byte		// uint32, little endian
// Message			[]byte		// data, msgpacked

type TorigoyaProtocolFrame struct {
	MessageKind		MessageKind	//
	Version			uint32		//
	Message			[]byte		// data, msgpacked
}

//
type ProtocolDataType map[string]string

//
func decodeToTorigoyaProtocol(reader io.Reader) (*TorigoyaProtocolFrame, error) {
	// read signature
	var sig [2]byte
	{
		n, err := reader.Read(sig[:])
		if err != nil { return nil, err }
		if n != 2 {
			return nil, errors.New("invalid signature length")
		}
		if sig != TorigoyaProtocolSignature {
			return nil, errors.New("invalid signature")
		}
	}

	// read kind
	var kind [1]uint8
	{
		n, err := reader.Read(kind[:])
		if err != nil { return nil, err }
		if n != 1 {
			return nil, errors.New("invalid kind length")
		}
	}

	// read version
	var version_bs [4]byte
	var version uint32
	{
		n, err := reader.Read(version_bs[:])
		if err != nil { return nil, err }
		if n != 4 {
			return nil, errors.New("invalid version length")
		}
		if err := binary.Read(
			bytes.NewReader(version_bs[:]),
			binary.LittleEndian,
			&version,
		); err != nil {
			return nil, err
		}
	}

	// read length
	var length_bs [4]byte
	var length uint32
	{
		n, err := reader.Read(length_bs[:])
		if err != nil { return nil, err }
		if n != 4 {
			return nil, errors.New("invalid length length")
		}
		if err := binary.Read(
			bytes.NewReader(length_bs[:]),
			binary.LittleEndian,
			&length,
		); err != nil {
			return nil, err
		}
	}

	// message length limit: 1MiB
	if length > 1 * 1024 * 1024 {
		return nil, errors.New("Message length limitation")
	}

	message := make([]byte, length)
	{
		n, err := io.ReadFull(reader, message[:length])
		if err != nil {
			return nil, err
		}
		if uint32(n) != length {
			return nil, errors.New(
				fmt.Sprintf("message length should be %d but got %d", length, n),
			)
		}
	}

	return &TorigoyaProtocolFrame{
		MessageKind: MessageKind(kind[0]),
		Version: version,
		Message: message,
	}, nil

	return nil, nil
}

//
func encodeToTorigoyaProtocol(
	writer	io.Writer,
	kind	MessageKind,
	version	uint32,
	object	interface{},
) error {
	// encode data
	var body_buffer []byte
	enc := codec.NewEncoderBytes(&body_buffer, &msgPackHandler)
	if err := enc.Encode(&object); err != nil {
		return err
	}

	// write signature(2Bytes)
	if err := binary.Write(
		writer,
		binary.LittleEndian,
		TorigoyaProtocolSignature,
	); err != nil { return err }

	// write kind(1Bytes)
	if kind < MessageKindIndexBegin || kind > MessageKindIndexEnd {
		return errors.New(fmt.Sprintf("Failed to write data / invalid header %d", kind))
	}
	if err := binary.Write(
		writer,
		binary.LittleEndian,
		kind,
	); err != nil { return err }

	// write version(4Bytes)
	if err := binary.Write(
		writer,
		binary.LittleEndian,
		version,
	); err != nil { return err }

	// length(4Bytes)
	length := uint32(len(body_buffer))
	if err := binary.Write(
		writer,
		binary.LittleEndian,
		length,
	); err != nil { return err }

	// data(length Bytes)
	n, err := writer.Write(body_buffer)
	if err != nil { return err }
	if uint32(n) != length {
		return errors.New("Failed to write data: length are different")
	}

	return nil
}
