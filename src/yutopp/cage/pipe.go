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
	"errors"
)


// MEMO: maybe, closed flag should not be copied...
type Pipe struct {
	ReadFd, WriteFd				int
	readClosed, writeClosed		bool
}

func makePipe() (*Pipe, error) {
	return makePipeWithFlags(0)
}

func makePipeCloseOnExec() (*Pipe, error) {
	return makePipeWithFlags(syscall.O_CLOEXEC)
}

func makePipeNonBlocking() (*Pipe, error) {
	p, err := makePipe()
	if err != nil { return p, err }

	if _, _, errno := syscall.RawSyscall(syscall.SYS_FCNTL, uintptr(p.ReadFd), syscall.F_SETFL, syscall.O_NONBLOCK); errno != 0 {
		return p, errors.New("")
	}

	return p, nil
}

func makePipeWithFlags(flags int) (*Pipe, error) {
	pipe := make([]int, 2)
	if err := syscall.Pipe2(pipe, flags); err != nil {
		return nil, err
	}

	return &Pipe{pipe[0], pipe[1], false, false}, nil
}

func (p *Pipe) CopyForClone() *Pipe {
	return &Pipe{p.ReadFd, p.WriteFd, false, false}
}

func (p *Pipe) Close() error {
	if err := p.CloseRead(); err != nil { return err }
	return p.CloseWrite()
}

func (p *Pipe) CloseRead() error {
	if !p.readClosed {
		if err := syscall.Close(p.ReadFd); err != nil { return err }
		p.readClosed = true
	}
	return nil
}

func (p *Pipe) CloseWrite() error {
	if !p.writeClosed {
		if err := syscall.Close(p.WriteFd); err != nil { return err }
		p.writeClosed = true
	}
	return nil
}
