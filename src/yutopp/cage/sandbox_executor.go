//
// Copyright yutopp 2015 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import(
	"os"
)

type MountOption struct {
	HostPath	string
	GuestPath	string
	IsReadOnly	bool
}

type CopyOption struct {
	HostPath	string
	GuestPath	string
}

type ResourceLimit struct {
	Core		uint64	// number
	Nofile		uint64	// number
	NProc		uint64	// number
	MemLock		uint64	// number
	CpuTime		uint64	// seconds
	Memory		uint64	// bytes
	FSize		uint64	// bytes
}

type SandboxExecutionOption struct {
	Mounts			[]MountOption
	Copies			[]CopyOption
	GuestHomePath	string
	Limits			*ResourceLimit
	Args			[]string
	Envs			[]string
}

type ExecuteCallBackType	func(*StreamOutput)
type SandboxExecutor interface {
	Execute(*SandboxExecutionOption, *os.File, ExecuteCallBackType) (*ExecutedResult, error)
}
