//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import(
	"github.com/ugorji/go/codec"
)


// Status of Result
const (
    MemoryLimit = 1
    CPULimit = 2
    OutputLimit = 22
    Error = 3
    InvalidCommand = 31
    Passed = 4
    UnexpectedError = 5
)

//
type ExecutedResult struct {
	UsedCPUTimeSec		float32
	UsedMemoryBytes		uint64
	Signal				int
	ReturnCode			int
	CommandLine			string
	IsSystemFailed		bool
	SystemErrorMessage	string
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
