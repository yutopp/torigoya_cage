package torigoya

import (
	"net"
	"time"
	"strconv"
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"
)


//
const ServerVersion = "v2014/7/5"

//
func RunServer(
	host string,
	port int,
	context *Context,
	notifier chan<-error,
	notify_pid int,
) error {
	laddr := makeAddress(host, port)
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		notifier <- err
		return err
	}
	defer listener.Close()

	// there are no error
	if notifier != nil { notifier <- nil }

	//
	if notify_pid != -1 {
		process, err := os.FindProcess(notify_pid)
		if err != nil {
			log.Panicf("Error (%v)\n", err)
		}
		if err := process.Signal(syscall.SIGUSR1); err != nil {
			log.Panicf("Error (%v)\n", err)
		}
	}

	for {
		// Wait for a connection.
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Server / Error: %v\n", err)
			continue
		}

		log.Printf("Server / Accepted: %v\n", conn)
		go handleConnection(conn, context)
	}

	return nil
}


func handleConnection(c net.Conn, context *Context) {
	var handler ProtocolHandler
	log.Printf("Server connection %v\n", c)

	//
	defer func() {
		if i := recover(); i != nil {
			if err, ok := i.(error); ok {
				log.Printf("handleConnection::Failed: %v\n", err)
				handler.writeSystemError(c, err.Error())
			}
        }

		// retry 5times if failed...
		for i:=0; i<5; i++ {
			if err := handler.writeExit(c); err == nil {
				// if error NOT returnd, ok
				break
			}
		}

		//
		c.Close()

		//
		log.Printf("Server connection CLOSED %v\n", c)
	}()

	//
	error_event := make(chan error)

	//
	go func() {
		defer close(error_event)

		if err := acceptGreeting(c, context, &handler, error_event); err != nil {
			return
		}

		acceptRequestMessage(c, context, &handler, error_event)

		log.Printf("Resuest passed\n")
		error_event <- nil
	}()

	// wait
	for err := range error_event {
		if err != nil {
			panic(err)
		}
	}
}


func acceptGreeting(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
	error_event chan<-error,
) error {
	// set timeout at the first time
	c.SetReadDeadline(time.Now().Add(10 * time.Second))

	//
	kind, data, err := handler.read(c)
	if err != nil {
		e := errors.New(fmt.Sprintf("Reciever error(%v)", err))
		error_event <- e
		return e
	}
	log.Printf("Server::Recieved: %d / %V\n", kind, data)

	// switch process by kind
	switch kind {
	case MessageKindAcceptRequest:
		version_bytes, ok := data.([]uint8)
		if !ok {
			e := errors.New("Given version data is not string")
			error_event <- e
			return e
		}
		version := string(version_bytes)
		log.Printf("Client Version : %s\n", version)

		// version matching
		if version != ServerVersion {
			e := errors.New(fmt.Sprintf("Client version is different from server (Server: %s / Client: %s)", ServerVersion, version))
			error_event <- e
			return e
		}

		// return accept message
		var err error = nil
		for i:=0; i<5; i++ {		// retry 5times if failed...
			if err = handler.writeAccept(c); err == nil {
				return nil
			}
		}
		e := errors.New("Failed to send output result : " + err.Error())
		error_event <- e
		return e

	default:
		e := errors.New("Server can accept only 'AcceptRequest' messages")
		error_event <- e
		return e
	}
}

func acceptRequestMessage(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
	error_event chan<-error,
) {
	// set timeout at the first time
	c.SetReadDeadline(time.Now().Add(10 * time.Second))

	//
	kind, data, err := handler.read(c)
	if err != nil {
		error_event <- errors.New(fmt.Sprintf("Reciever error(%v)", err))
		return
	}
	log.Printf("Server::Recieved: %d / %V\n", kind, data)

	// switch process by kind
	switch kind {
	case MessageKindTicketRequest:
		// accept ticket execution request
		acceptTicketRequestMessage(data, c, context, handler, error_event)

	case MessageKindUpdateRepositoryRequest:
		// install/upgrade APT repository
		acceptUpdateRepositoryRequest(c, context, handler, error_event)

	case MessageKindReloadProcTableRequest:
		// reload ProcProfiles
		acceptReloadProcTableRequest(c, context, handler, error_event)

	case MessageKindUpdateProcTableRequest:
		// update ProcProfiles
		acceptUpdateProcTableRequest(c, context, handler, error_event)

	case MessageKindGetProcTableRequest:
		// send ProcProfiles to the client
		acceptGetProcTableMessage(c, context, handler, error_event)

	default:
		error_event <- errors.New(fmt.Sprintf("Server can not accept message (%d)", kind))
		return
	}
}

//
func acceptTicketRequestMessage(
	data interface{},
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
	error_event chan<-error,
) {
	// execute ticket
	fmt.Printf("ticket request %V\n", data)
	ticket, err := MakeTicketFromTuple(data)
	if err != nil {
		fmt.Printf("Server::Invalid request (%s)\n", err.Error())
		error_event <- errors.New(fmt.Sprintf("Invalid request (%s)", err.Error()))
		return
	}
	fmt.Printf("ticket %V\n", ticket)

	// callback function
	error_happend := false
	f := func(v interface{}) {
		if error_happend { return }

		switch v.(type) {
		case *StreamOutputResult:
			var err error = nil
			for i:=0; i<5; i++ {		// retry 5times if failed...
				if err = handler.writeOutputResult(c, v.(*StreamOutputResult)); err == nil {
					return
				}
			}
			error_event <- errors.New("Failed to send output result : " + err.Error())
			error_happend = true
			return

		case *StreamExecutedResult:
			var err error = nil
			for i:=0; i<5; i++ {		// retry 5times if failed...
				if err = handler.writeExecutedResult(c, v.(*StreamExecutedResult)); err == nil {
					return
				}
			}
			error_event <- errors.New("Failed to send executed result : " + err.Error())
			error_happend = true
			return

		default:
			error_event <- errors.New("Unsupported type object was given to callback")
			error_happend = true
			return
		}
	}

	// execute ticket data
	if err := context.ExecTicket(ticket, f); err != nil {
		fmt.Printf("Server::Failed to exec ticket (%s)\n", err.Error())
		error_event <- errors.New(fmt.Sprintf("Failed to exec ticket (%s)", err.Error()))
		return
	}
}

//
func acceptUpdateRepositoryRequest(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
	error_event chan<-error,
) {
	if err := context.UpdatePackages(); err != nil {
		error_event <- err
		return
	}

	var err error = nil
	for i:=0; i<5; i++ {		// retry 5times if failed...
		if err = handler.writeSystemResult(c, 0); err == nil {
			return
		}
	}

	error_event <- errors.New("Failed to send system request: " + err.Error())
}

//
func acceptReloadProcTableRequest(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
	error_event chan<-error,
) {
	if err := context.ReloadProcTable(); err != nil {
		error_event <- err
		return
	}

	var err error = nil
	for i:=0; i<5; i++ {		// retry 5times if failed...
		if err = handler.writeSystemResult(c, 0); err == nil {
			return
		}
	}

	error_event <- errors.New("Failed to send system request: " + err.Error())
}

//
func acceptUpdateProcTableRequest(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
	error_event chan<-error,
) {
	if err := context.UpdateProcTable(); err != nil {
		error_event <- err
		return
	}

	var err error = nil
	for i:=0; i<5; i++ {		// retry 5times if failed...
		if err = handler.writeSystemResult(c, 0); err == nil {
			return
		}
	}

	error_event <- errors.New("Failed to send system request: " + err.Error())
}

//
func acceptGetProcTableMessage(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
	error_event chan<-error,
) {
	var err error = nil
	for i:=0; i<5; i++ {		// retry 5times if failed...
		if err = handler.writeProcTable(c, &context.procConfTable); err == nil {
			return
		}
	}

	error_event <- errors.New("Failed to send proc table: " + err.Error())
}


//
func makeAddress(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}
