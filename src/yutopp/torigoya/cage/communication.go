package torigoya

import (
	"net"
	"time"
	"strconv"
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
	//
	defer c.Close()

	//
	var handler ProtocolHandler

	// set timeout at the first time
	c.SetReadDeadline(time.Now().Add(10 * time.Second))
	kind, data, err := handler.Read(c)
	if err != nil {
		handler.WriteError(c, fmt.Sprintf("Reciever error(%v)", err))
		return
	}

	// switch process by kind
	switch kind {
	case HeaderRequest:
		if err := handler.WriteRequest(c, "aba"); err != nil {
			log.Printf("Server / Error: %v\n", err)
		}
		fmt.Printf("data %V\n", data)
		ticket, err := MakeTicketFromTuple(data)
		if err != nil {
			fmt.Printf("dame %V\n", err)
		}
		fmt.Printf("ticket %V\n", ticket)

	default:
		handler.WriteError(c, fmt.Sprintf("Server can accept only 'Request' messages"))
	}

	log.Printf("Server recv: %d / %v\n", kind, data)
}
