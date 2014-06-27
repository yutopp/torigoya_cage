//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

// +build linux

package torigoya

import(
	"log"
	"errors"

	"encoding/base64"
	"github.com/ugorji/go/codec"
)


//
type BridgePipes struct {
	Stdout, Stderr, Result	*Pipe
}

func (bp *BridgePipes) Close() {
	bp.Stdout.Close()
	bp.Stderr.Close()
	bp.Result.Close()
}


//
type BridgeMessage struct {
	ChrootPath			string
	JailedUserHomePath	string
	JailedUser			*JailedUserInfo
	Pipes				*BridgePipes
	Message				ExecMessage
	IsReboot			bool
}

func (bm *BridgeMessage) Encode() (string, error) {
	var msgpack_bytes []byte
	enc := codec.NewEncoderBytes(&msgpack_bytes, &msgPackHandler)
	if err := enc.Encode(*bm); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(msgpack_bytes), nil
}

func DecodeBridgeMessage(base string) (*BridgeMessage, error) {
	decoded_bytes, err := base64.StdEncoding.DecodeString(base)
	if err != nil {
		return nil, err
	}

	bm := &BridgeMessage{}
	dec := codec.NewDecoderBytes(decoded_bytes, &msgPackHandler)
	if err := dec.Decode(bm); err != nil {
		return nil, err
	}

	return bm, nil
}


//
func (bm *BridgeMessage) Exec() error {
	m := bm.Message

	exec_result, err := func() (*ExecutedResult, error) {
		switch m.Mode {
		case CompileMode:
			return bm.compile()
		case LinkMode:
			return bm.link()
		case RunMode:
			return bm.run()
		default:
			return nil, errors.New("Exec:: Invalid mode")
		}
	}()
	if err != nil {
		exec_result = &ExecutedResult{
			Status: UnexpectedError,
			SystemErrorMessage: err.Error(),
		}
	}
	if exec_result == nil {
		exec_result = &ExecutedResult{
			Status: UnexpectedError,
			SystemErrorMessage: "Result was not generated",
		}
	}

	return exec_result.sendTo(bm.Pipes)
}

//
func (bm *BridgeMessage) compile() (*ExecutedResult, error) {
	log.Println(">> called BridgeMessage::compile")
	exec_message := bm.Message

	proc_profile := exec_message.Profile
	stdin_file_path := exec_message.StdinFilePath
	exec_setting := exec_message.Setting

	// arguments
	args, err := proc_profile.Compile.MakeCompleteArgs(
		exec_setting.CommandLine,
		exec_setting.StructuredCommand,
	)
	if err != nil {
		return nil, err
	}

	//
	res_limit := &ResourceLimit{
		CPU: exec_setting.CpuTimeLimit,		// CPU limit(sec)
		AS: exec_setting.MemoryBytesLimit,	// Memory limit(bytes)
		FSize: 5 * 1024 * 1024,				// Process can writes a file only 5 MBytes
	}

	_ = stdin_file_path

	return managedExec(res_limit, bm.Pipes, args, map[string]string{"PATH": "/bin"})
}

func (bm *BridgeMessage) link() (*ExecutedResult, error) {
	log.Println(">> called BridgeMessage::link")

	return nil, nil
}

func (bm *BridgeMessage) run() (*ExecutedResult, error) {
	log.Println(">> called BridgeMessage::run")

	return nil, nil
}
