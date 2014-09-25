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
	"fmt"
	"syscall"
	"errors"
)


// MEMO: maybe, closed flag should not be copied...
type Pipe struct {
	ReadFd, WriteFd				int
	readClosed, writeClosed		bool
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

func (p *Pipe) Dup() (*Pipe, error) {
	var err error

	var new_readfd int = -1
	if !p.readClosed {
		new_readfd, err = syscall.Dup(p.ReadFd)
		if err != nil { return nil, err }
	}

	var new_writefd int = -1
	if !p.writeClosed {
		new_writefd, err = syscall.Dup(p.WriteFd)
		if err != nil { return nil, err }
	}

	return &Pipe{new_readfd, new_writefd, p.readClosed, p.writeClosed}, nil
}

func (p *Pipe) ToCloseOnExec() (error) {
	if ! p.readClosed {
		if _, _, errno := syscall.RawSyscall(syscall.SYS_FCNTL, uintptr(p.ReadFd), syscall.F_SETFD, syscall.FD_CLOEXEC); errno != 0 {
			return errors.New(fmt.Sprintf("Failed ToCloseOnExec(Read): %d", errno))
		}
	}

	if ! p.writeClosed {
		if _, _, errno := syscall.RawSyscall(syscall.SYS_FCNTL, uintptr(p.WriteFd), syscall.F_SETFD, syscall.FD_CLOEXEC); errno != 0 {
			return errors.New(fmt.Sprintf("Failed ToCloseOnExec(Write): %d", errno))
		}
	}

	return nil
}

// ================================================================================

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

func makePipeNonBlockingWithFlags(flags int) (*Pipe, error) {
	p, err := makePipeWithFlags(flags)
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
