package torigoya

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ugorji/go/codec"
)

//
const ServerVersion = uint32(20150715)
const channelBuffer = 2048

//
func RunServer(
	host string,
	port int,
	context *Context,
	notifier chan<- error,
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
			os.Exit(-1)
		}
	}()

	// there is no error
	notifier <- nil

	for {
		// Wait for a connection.
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Server Error[Accept]: %v\n", err)
			continue
		}

		log.Printf(" I  Server Accepted: %v\n", conn)
		go handleConnection(conn, context)
	}

	return nil
}

func retryIfFailed(f func() error) (err error) {
	// retry 5times if failed...
	for i := 0; i < 5; i++ {
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
	log.Printf("[+] Server Connection %v\n", c)

	handler := &ProtocolHandler{
		Io:      c,
		Version: ServerVersion,
	}

	defer func() {
		if i := recover(); i != nil {
			if err, ok := i.(error); ok {
				log.Printf("handleConnection::Failed: %v\n", err)
				handler.writeSystemError(err.Error())
			}
		}

		_ = retryIfFailed(func() error {
			return handler.writeExit(c)
		})

		c.Close()
		log.Printf("[-] Server Connection CLOSED %v\n", c)
	}()

	// set read timeout at the first time
	c.SetReadDeadline(time.Now().Add(10 * time.Second))

	if err := acceptRequestMessage(context, handler); err != nil {
		log.Printf("    Resuest failed: %v / %v\n", err, c)
		handler.writeSystemError(err.Error())
		return
	}

	log.Printf("    Resuest passed %v\n", c)
}

func acceptRequestMessage(
	context *Context,
	handler *ProtocolHandler,
) error {
	kind, buffer, err := handler.read()
	if err != nil {
		return errors.New(fmt.Sprintf("Reciever error at acceptRequestMessage(%V)", err))
	}

	// log.Printf("Server::Recieved: %s / %V\n", kind.String(), buffer)

	// switch process by kind
	switch kind {
	case MessageKindTicketRequest:
		// accept ticket execution request
		return acceptTicketRequestMessage(buffer, context, handler)

	case MessageKindUpdateRepositoryRequest:
		// install/upgrade APT repository
		return acceptUpdateRepositoryRequest(context, handler)

	default:
		return errors.New(fmt.Sprintf("Server can not accept message (%d)", kind))
	}
}

//
type term struct{}

func acceptTicketRequestMessage(
	buffer []byte,
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
	// log.Printf("ticket %V\n", ticket)

	// callback function
	error_happend := false
	var comm_err error = nil
	results_ch := make(chan interface{}, channelBuffer)
	f := func(v interface{}) error {
		if error_happend {
			log.Printf("ERROR CALLBACK: %v", v)
			return errors.New("error flag")
		}

		select {
		case results_ch <- v: // push value
		case <-time.After(3 * time.Second):
			return errors.New("callback timeout")
		}

		return nil
	}

	reading_ch := make(chan error, 10)
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
				log.Printf("in:  StreamOutputResult / %v", v.(*StreamOutputResult))

				if err := retryIfFailed(func() error {
					return handler.writeOutputResult(v.(*StreamOutputResult))
				}); err != nil {
					log.Printf("StreamOutputResult / Error >> %v", err)
					reading_ch <- err
					error_happend = true
					return
				}
				log.Printf("out: StreamOutputResult / %v", v.(*StreamOutputResult))

			case *StreamExecutedResult:
				log.Printf("in:  StreamExecutedResult / %v", v.(*StreamExecutedResult))

				if err := retryIfFailed(func() error {
					return handler.writeExecutedResult(v.(*StreamExecutedResult))
				}); err != nil {
					log.Printf("StreamExecutedResult / Error >> %v", err)
					reading_ch <- err
					error_happend = true
					return
				}
				log.Printf("out: StreamExecutedResult / %v", v.(*StreamExecutedResult))

			case *term:
				log.Println("in:  terminates")
				reading_ch <- nil
				log.Println("out: terminates")
				return

			default:
				log.Println("in:  error")
				err := errors.New("Unsupported type object was given to callback")
				reading_ch <- err
				error_happend = true
				log.Println("out: error")
				return
			}
		}

		// TODO: make it error
		//reading_ch <- nil
	}()

	// execute ticket data
	if err := context.ExecTicket(&ticket, f); err != nil {
		fmt.Printf("Server::Failed to exec ticket (%s)\n", err.Error())
		return fmt.Errorf("Failed to exec ticket (%s)", err.Error())
	}
	results_ch <- &term{}

	<-reading_ch

	return comm_err
}

//
func acceptUpdateRepositoryRequest(
	context *Context,
	handler *ProtocolHandler,
) error {
	if err := context.UpdatePackages(); err != nil {
		return err
	}

	if err := retryIfFailed(func() error {
		return handler.writeSystemResult(0)
	}); err != nil {
		return errors.New("Failed to send system request: " + err.Error())
	}

	return nil
}

//
func makeAddress(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}
