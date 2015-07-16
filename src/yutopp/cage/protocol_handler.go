//
// Copyright yutopp 2015 - .
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
	"errors"
)

type ProtocolHandler struct {
	Io			io.ReadWriter
	Version		uint32
}

func (ph *ProtocolHandler) read() (MessageKind, []byte, error) {
	// read protocol
	frame, err := decodeToTorigoyaProtocol(ph.Io)
	if err != nil {
		return MessageKindInvalid, nil, err
	}

	if frame.Version != ph.Version {
		return MessageKindInvalid, nil, errors.New(
			fmt.Sprintf(
				"Protocol version is different (Require: %s / Actual: %s)",
				ph.Version, frame.Version,
			),
		)
	}
	return frame.MessageKind, frame.Message, nil
}

func (ph *ProtocolHandler) write(
	kind MessageKind,
	object interface{},
) error {
	res_writer := bytes.NewBuffer(nil)
	if err := encodeToTorigoyaProtocol(
		res_writer,
		kind,
		ph.Version,
		object,
	); err != nil {
		return err
	}

	buffer := res_writer.Bytes()
	n, err := ph.Io.Write(buffer)
	if err != nil {
		return err
	}
	if n != len(buffer) {
		return errors.New("couldn't send all bytes")
	}

	return nil
}

func (ph *ProtocolHandler) writeOutputResult(
	r *StreamOutputResult,
) error {
	return ph.write(MessageKindOutputs, r)
}

func (ph *ProtocolHandler) writeExecutedResult(
	r *StreamExecutedResult,
) error {
	return ph.write(MessageKindResult, r)
}

func (ph *ProtocolHandler) writeSystemError(
	message string,
) error {
	return ph.write(MessageKindSystemError, message)
}

func (ph *ProtocolHandler) writeExit(
	writer io.Writer,
) error {
	return ph.write(MessageKindExit, "")
}

func (ph *ProtocolHandler) writeSystemResult(
	status int,
) error {
	return ph.write(MessageKindSystemResult, status)
}
