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
	"syscall"
)

// MEMO: maybe, closed flag should not be copied...
type Pipe struct {
	ReadFd, WriteFd				int
	readClosed, writeClosed		bool
}

func makePipe() (*Pipe, error) {
	pipe := make([]int, 2)
	if err := syscall.Pipe(pipe); err != nil {
		return nil, err
	}

	return &Pipe{pipe[0], pipe[1], false, false}, nil
}

func (p *Pipe) Close() {
	p.CloseRead()
	p.CloseWrite()
}

func (p *Pipe) CloseRead() {
	if !p.readClosed {
		syscall.Close(p.ReadFd)
		p.readClosed = true
	}
}

func (p *Pipe) CloseWrite() {
	if !p.writeClosed {
		syscall.Close(p.WriteFd)
		p.writeClosed = true
	}
}
