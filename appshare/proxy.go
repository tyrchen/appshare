package appshare

import (
	"bufio"
	"bytes"
	"github.com/op/go-logging"
	"io"
	"net"
)

const (
	READ_BUFFER = 8192
	MAXSUITORS  = 64
)

var (
	log = logging.MustGetLogger("appshare")
)

// Connection is the client connection to the Proxy. It is created when the
// listener accepts the conneciton.
type Connection struct {
	conn     net.Conn      // the connection
	incoming chan []byte   // channel for incoming data of the connection
	outgoing chan []byte   // channel for outgoing data of the connection
	reader   *bufio.Reader // reader to the connection
	writer   *bufio.Writer // writer to the connection
	quiting  chan net.Conn // channel for gracefully quit the connection
}

type ConnectionTable map[net.Conn]*Connection

// Proxy is the common data structure for proxy requests/response with web
// servers.

type Proxy struct {
	listener net.Listener    // listener to the connection
	clients  ConnectionTable // holding the connections from client side
	servers  ConnectionTable // holding the connectiosn from server side
	pending  chan net.Conn   // channel for accepted connection
	quiting  chan net.Conn   // channel for connection being quit
	//terminate chan bool     // channel for terminate itself
	incoming chan []byte // channel for incoming data of the proxy
	outgoing chan []byte // channel for outgoing data of the proxy

}

// Methods for Connection

func CreateConnection(conn net.Conn) *Connection {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	connection := &Connection{
		conn:     conn,
		incoming: make(chan []byte),
		outgoing: make(chan []byte),
		quiting:  make(chan net.Conn),
		reader:   reader,
		writer:   writer,
	}
	connection.Listen()
	return connection
}

func (self *Connection) Listen() {
	go self.Read()
	go self.Write()
}

func (self *Connection) quit() {
	close(self.quiting)
}

func (self *Connection) Read() {
	data := make([]byte, READ_BUFFER)
	// keep reading or pending on self.incoming
	for {
		select {
		case <-self.quiting:
			return
		default:
			total := 0
			buf := &bytes.Buffer{}
			for {
				n, err := self.reader.Read(buf)
				total += n
				buf.Write(data[:n])
				if err != nil {
					if err != io.EOF {
						log.Critical("Read error: %s", err)
						self.quit()
					}
					break
				}
			}
			log.Info("%d bytes read", total)
			self.incoming <- buf.Bytes()
		}
	}

}

func (self *Connection) Write() {
	for {
		select {
		case <-self.quiting:
			return
		case data := <-self.outgoing:
			if _, err := self.writer.Write(data); err != nil {
				log.Critical("Write error: %s", err)
				self.quit()
			}

			if err := self.writer.Flush(); err != nil {
				log.Critical("Flush error: %s", err)
				self.quit()
			}
		}
	}

}

func (self *Connection) Close() {
	self.conn.Close()
}

// Methods for Proxy

func CreateProxy() *Proxy {
	proxy := &Proxy{
		clients:  make(ConnectionTable),
		servers:  make(ConnectionTable),
		pending:  make(chan net.Conn),
		quiting:  make(chan net.Conn),
		incoming: make(chan []byte),
		outgoing: make(chan []byte),
	}
	proxy.listen()
	return proxy
}

func (self *Proxy) listen() {
	go func() {
		for {
			select {
			case conn := <-self.pending:
				self.accept(conn)
			case conn := <-self.quiting:
				self.cleanup(conn)
			}
		}
	}()
}

func (self *Proxy) accept(conn net.Conn) {
	connection := CreateConnection(conn)
	self.clients[conn] = connection

	log.Info("A new connection %v kicks in.", conn)

	go func() {
		for {
			select {
			case conn := <-connection.quiting:
				self.quiting <- conn
				return
			case data := connection.incoming:
				log.Info("Got data from client %v\n", connection)
				self.incoming <- data
			}
		}
	}()
}

func (self *Proxy) cleanup(conn net.Conn) {
	if conn != nil {
		connection := self.connections[conn]
		connection.Close()
		delete(self.clients, conn)
	}
}

func (self *Proxy) DoProxy(outgoing, incoming chan []byte) {
	go func() {
		for {
			select {
			case data := <-self.incoming:
				log.Info("Proxy data out")
				log.Debug(string(data))
				outgoing <- data
			case data := incoming:
				log.Info("Proxy data in")
				log.Debug(string(data))
				self.outgoing <- incoming
			}
		}
	}()
}

func (self *Proxy) Start(connString string) {
	self.listener, _ = net.Listen("tcp", connString)

	log.Notice("Proxy server %p starts\n", self)

	for {
		conn, err := self.listener.Accept()

		if err != nil {
			log.Error(err)
			continue
		}

		log.Info("A new connection %v kicks\n", conn)
		self.pending <- conn
	}
}
