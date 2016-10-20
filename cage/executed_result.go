//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

type OutFd int

const (
	// StdinFd = OutFd(0)
	StdoutFd = OutFd(1)
	StderrFd = OutFd(2)
)

type StreamOutput struct {
	Fd     OutFd  `codec:"fd"`
	Buffer []byte `codec:"buffer"`
}

type StreamOutputResult struct {
	Mode      int           `codec:"mode"`
	MainIndex int           `codec:"main_index"`
	SubIndex  int           `codec:"sub_index"`
	Output    *StreamOutput `codec:"output"`
}

type StreamExecutedResult struct {
	Mode      int             `codec:"mode"`
	MainIndex int             `codec:"main_index"`
	SubIndex  int             `codec:"sub_index"`
	Result    *ExecutedResult `codec:"result"`
}

type ExecutedResult struct {
	Exited     bool `codec:"exited"`
	ExitStatus int  `codec:"exit_status"`
	Signaled   bool `codec:"signaled"`
	Signal     int  `codec:"signal"`

	UsedCPUTimeSec  float64 `codec:"used_cpu_time_sec"`
	UsedMemoryBytes uint64  `codec:"used_memory_bytes"`

	SystemErrorStatus  int    `codec:"system_error_status"`
	SystemErrorMessage string `codec:"system_error_message"`
}

func (bm *ExecutedResult) IsSucceeded() bool {
	return bm.SystemErrorStatus == 0 && bm.Exited && bm.ExitStatus == 0
}
