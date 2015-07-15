package torigoya

import (
	"net"
	"time"
	"strconv"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ugorji/go/codec"
)


//
const ServerVersion = "20150715"

//
func RunServer(
	host string,
	port int,
	context *Context,
	notifier chan<-error,
	notify_pid int,
) error {
	if notifier == nil {
		return errors.New("notifier must be specified")
	}

	laddr := makeAddress(host, port)
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		notifier <- err
		return err
	}
	defer listener.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for _ = range c {
			log.Printf("Signal captured\n")
			listener.Close()
			os.Exit(0)
		}
	}()

	// there is no error
	notifier <- nil

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
			log.Printf("Server Error[Accept]: %v\n", err)
			continue
		}

		log.Printf("Server Accepted: %v\n", conn)
		go handleConnection(conn, context)
	}

	return nil
}

func retryIfFailed(f func() error) (err error) {
	// retry 5times if failed...
	for i:=0; i<5; i++ {
		if err = f(); err == nil {
			// if there is no error
			return nil
		}
	}

	return
}

type session struct {
}

func handleConnection(c net.Conn, context *Context) {
	var handler ProtocolHandler
	log.Printf("Server connection %v\n", c)

	defer func() {
		if i := recover(); i != nil {
			if err, ok := i.(error); ok {
				log.Printf("handleConnection::Failed: %v\n", err)
				handler.writeSystemError(c, err.Error())
			}
        }

		_ = retryIfFailed(func() error {
			return handler.writeExit(c)
		})

		c.Close()
		log.Printf("Server connection CLOSED %v\n", c)
	}()

	//
	if err := acceptGreeting(c, context, &handler); err != nil {
		panic(err)
	}

	if err := acceptRequestMessage(c, context, &handler); err != nil {
		panic(err)
	}

	log.Printf("Resuest passed %v\n", c)
}


func acceptGreeting(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
) error {
	// set timeout at the first time
	c.SetReadDeadline(time.Now().Add(10 * time.Second))

	//
	log.Printf("acceptGreeting\n")
	kind, buffer, err := handler.read(c)
	if err != nil {
		e := errors.New(fmt.Sprintf("Reciever error at Greeting(%V)", err))
		return e
	}
	log.Printf("Server::Recieved: %s / %V\n", kind.String(), buffer)

	// switch process by kind
	switch kind {
	case MessageKindAcceptRequest:
		// decode
		var version string
		dec := codec.NewDecoderBytes(buffer, &msgPackHandler)
		if err := dec.Decode(&version); err != nil {
			return err
		}

		log.Printf("Client Version : %s\n", version)

		// version matching
		if version != ServerVersion {
			e := errors.New(fmt.Sprintf("Client version is different from server (Server: %s / Client: %s)", ServerVersion, version))
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
		return e

	default:
		e := errors.New("Server can accept only 'AcceptRequest' messages")
		return e
	}
}

func acceptRequestMessage(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
) error {
	// set timeout at the first time
	c.SetReadDeadline(time.Now().Add(10 * time.Second))

	//
	kind, buffer, err := handler.read(c)
	if err != nil {
		return errors.New(fmt.Sprintf("Reciever error at acceptRequestMessage(%V)", err))
	}
	log.Printf("Server::Recieved: %s / %V\n", kind.String(), buffer)

	// switch process by kind
	switch kind {
	case MessageKindTicketRequest:
		// accept ticket execution request
		return acceptTicketRequestMessage(buffer, c, context, handler)

	case MessageKindUpdateRepositoryRequest:
		// install/upgrade APT repository
		return acceptUpdateRepositoryRequest(c, context, handler)

	case MessageKindReloadProcTableRequest:
		// reload ProcProfiles
		return acceptReloadProcTableRequest(c, context, handler)

	case MessageKindUpdateProcTableRequest:
		// update ProcProfiles
		return acceptUpdateProcTableRequest(c, context, handler)

	case MessageKindGetProcTableRequest:
		// send ProcProfiles to the client
		return acceptGetProcTableMessage(c, context, handler)

	default:
		return errors.New(fmt.Sprintf("Server can not accept message (%d)", kind))
	}
}

type hoge struct {}

//
func acceptTicketRequestMessage(
	buffer []byte,
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
) error {
	log.Printf(">> begin acceptTicketRequestMessage")
	defer log.Printf("<< exit acceptTicketRequestMessage")

	var ticket Ticket
	dec := codec.NewDecoderBytes(buffer, &msgPackHandler)
	if err := dec.Decode(&ticket); err != nil {
		return err
	}

	// execute ticket
	log.Printf("ticket %V\n", ticket)

	// callback function
	error_happend := false
	var comm_err error = nil
	results_ch := make(chan interface{}, 100)
	f := func(v interface{}) {
		log.Printf("CALLBACK: %v", v)
		if error_happend { return }
		results_ch <- v
	}

	reading_ch := make(chan error)
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				reading_ch <- err.(error)
			}
        }()

		for v := range results_ch {
			switch v.(type) {
			case *StreamOutputResult:
				log.Printf("StreamOutputResult >> %v", v.(*StreamOutputResult))

				if err := retryIfFailed(func() error {
					return handler.writeOutputResult(c, v.(*StreamOutputResult))
				}); err != nil {
					log.Printf("StreamOutputResult / Error >> %v", err)
					reading_ch <- err
					error_happend = true
					return
				}
				break

			case *StreamExecutedResult:
				log.Printf("StreamExecutedResult >> %v", v.(*StreamExecutedResult))

				if err := retryIfFailed(func() error {
					return handler.writeExecutedResult(c, v.(*StreamExecutedResult))
				}); err != nil {
					log.Printf("StreamExecutedResult / Error >> %v", err)
					reading_ch <- err
					error_happend = true
					return
				}
				break

			case *hoge:
				reading_ch <- nil
				return

			default:
				err := errors.New("Unsupported type object was given to callback")
				reading_ch <- err
				error_happend = true
				return
			}
		}

		// TODO: make it error
		reading_ch <- nil
	}()

	// execute ticket data
	if err := context.ExecTicket(&ticket, f); err != nil {
		fmt.Printf("Server::Failed to exec ticket (%s)\n", err.Error())
		return errors.New(fmt.Sprintf("Failed to exec ticket (%s)", err.Error()))
	}
	results_ch <- &hoge{}

	<-reading_ch

	return comm_err
}

//
func acceptUpdateRepositoryRequest(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
) error {
	if err := context.UpdatePackages(); err != nil {
		return err
	}

	if err := retryIfFailed(func() error {
		return handler.writeSystemResult(c, 0)
	}); err != nil {
		return errors.New("Failed to send system request: " + err.Error())
	}

	return nil
}

//
func acceptReloadProcTableRequest(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
) error {
	if err := context.ReloadProcTable(); err != nil {
		return err
	}

	if err := retryIfFailed(func() error {
		return handler.writeSystemResult(c, 0)
	}); err != nil {
		return errors.New("Failed to send system request: " + err.Error())
	}

	return nil
}

//
func acceptUpdateProcTableRequest(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
) error {
	if err := context.UpdateProcTable(); err != nil {
		return err
	}

	if err := retryIfFailed(func() error {
		return handler.writeSystemResult(c, 0)
	}); err != nil {
		return errors.New("Failed to send system request: " + err.Error())
	}

	return nil
}

//
func acceptGetProcTableMessage(
	c net.Conn,
	context *Context,
	handler *ProtocolHandler,
) error {
	if err := retryIfFailed(func() error {
		return handler.writeProcTable(c, &context.procConfTable)
	}); err != nil {
		return errors.New("Failed to send proc table: " + err.Error())
	}

	return nil
}


//
func makeAddress(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}
