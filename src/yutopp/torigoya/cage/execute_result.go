//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import(
	"syscall"

	"github.com/ugorji/go/codec"
)


// Status of Result
type ExecutedStatus		int
const (
    MemoryLimit		= ExecutedStatus(1)
    CPULimit		= ExecutedStatus(2)
    OutputLimit		= ExecutedStatus(22)
    Error			= ExecutedStatus(3)
    InvalidCommand	= ExecutedStatus(31)
    Passed			= ExecutedStatus(4)
    UnexpectedError	= ExecutedStatus(5)
)

//
type ExecutedResult struct {
	UsedCPUTimeSec		float32
	UsedMemoryBytes		uint64
	Signal				*syscall.Signal
	ReturnCode			int
	CommandLine			string
	Status				ExecutedStatus
	SystemErrorMessage	string
}

func (bm *ExecutedResult) IsFailed() bool {
	return bm.Status != Passed;
}

func (bm *ExecutedResult) Encode() ([]byte, error) {
	var msgpack_bytes []byte
	enc := codec.NewEncoderBytes(&msgpack_bytes, &msgPackHandler)
	if err := enc.Encode(*bm); err != nil {
		return nil, err
	}
	return msgpack_bytes, nil
}

func DecodeExecuteResult(base []byte) (*ExecutedResult, error) {
	bm := &ExecutedResult{}
	dec := codec.NewDecoderBytes(base, &msgPackHandler)
	if err := dec.Decode(bm); err != nil {
		return nil, err
	}

	return bm, nil
}
