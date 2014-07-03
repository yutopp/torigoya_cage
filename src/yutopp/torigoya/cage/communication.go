package torigoya

import (
	"net"
	"time"
	"strconv"
	"errors"
	"fmt"
	"log"
)

func makeAddress(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}

func RunServer(
	host string,
	port int,
	context *Context,
	notifier chan<-error,
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
		//
		if i := recover(); i != nil {
			if err, ok := i.(error); ok {
				log.Printf("handleConnection::Failed: %v\n", err)
				handler.WriteSystemError(c, err.Error())
			}
        }

		// retry 5times if failed...
		for i:=0; i<5; i++ {
			if err := handler.WriteExit(c); err == nil {
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
		// set timeout at the first time
		c.SetReadDeadline(time.Now().Add(10 * time.Second))
		kind, data, err := handler.read(c)
		if err != nil {
			error_event <- errors.New(fmt.Sprintf("Reciever error(%v)", err))
		}
		log.Printf("Server recv: %d / %v\n", kind, data)

		//
		defer close(error_event)

		// switch process by kind
		switch kind {
		case HeaderRequest:
			fmt.Printf("Server::Recieved %V\n", data)
			ticket, err := MakeTicketFromTuple(data)
			if err != nil {
				fmt.Printf("Server::Invalid request (%s)\n", err.Error())
				error_event <- errors.New(fmt.Sprintf("Invalid request (%s)", err.Error()))
				return
			}
			fmt.Printf("ticket %V\n", ticket)

			f := func(v interface{}) {
				switch v.(type) {
				case *StreamOutputResult:
					var err error = nil
					for i:=0; i<5; i++ {		// retry 5times if failed...
						if err = handler.WriteOutputResult(c, v.(*StreamOutputResult)); err == nil {
							return
						}
					}
					error_event <- errors.New("Failed to send output result : " + err.Error())
					return

				case *StreamExecutedResult:
					var err error = nil
					for i:=0; i<5; i++ {		// retry 5times if failed...
						if err = handler.WriteExecutedResult(c, v.(*StreamExecutedResult)); err == nil {
							return
						}
					}
					error_event <- errors.New("Failed to send executed result : " + err.Error())
					return

				default:
					error_event <- errors.New("Unsupported type object was given to callback")
					return
				}
			}

			// execute ticket data
			if err := context.ExecTicket(ticket, f); err != nil {
				fmt.Printf("Server::Failed to exec ticket (%s)\n", err.Error())
				error_event <- errors.New(fmt.Sprintf("Failed to exec ticket (%s)", err.Error()))
				return
			}

		default:
			error_event <- errors.New("Server can accept only 'Request' messages")
			return
		}

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
