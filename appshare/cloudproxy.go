package appshare

import (
	//"bufio"
	"bytes"
	"io"
	"net"
)

const (
	READ_BUFFER = 8192
)

type TempData struct {
	conn net.Conn
	data []byte
}

type ConnPair struct {
	local  net.Conn // the connection
	remote net.Conn // the remote connection this will proxy to
}

// Connection is the client connection to the Proxy. It is created when the
// listener accepts the conneciton.
type Connection struct {
	pair     ConnPair
	incoming chan []byte // channel for incoming data of the connection
	//outgoing chan []byte   // channel for outgoing data of the connection
	//reader   *bufio.Reader // reader to the connection
	//writer   *bufio.Writer // writer to the connection
	quiting chan net.Conn // channel for gracefully quit the connection
}

type ConnectionTable map[net.Conn]*Connection

// Proxy is the common data structure for proxy requests/response with web
// servers.
type Proxy struct {
	listener net.Listener    // listener to the client connection
	clients  ConnectionTable // holding the connections from client side
	pending  chan net.Conn   // channel for accepted connection
	quiting  chan net.Conn   // channel for connection being quit
	//terminate chan bool     // channel for terminate itself
	incoming chan TempData // channel for holding incoming data which have no pair
	connPair chan ConnPair
}

// Methods for Connection

func CreateConnection(conn net.Conn) *Connection {
	//reader := bufio.NewReader(conn)
	//writer := bufio.NewWriter(conn)

	connection := &Connection{
		incoming: make(chan []byte),
		//outgoing: make(chan []byte),
		quiting: make(chan net.Conn),
		//reader:   reader,
		//writer:   writer,
	}
	connection.pair.local = conn
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
			if self.pair.remote != nil {
				io.Copy(self.pair.remote, self.pair.local)
			} else {
				total := 0
				buf := &bytes.Buffer{}
				for {
					n, err := self.pair.local.Read(data)
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

}

func (self *Connection) Write() {
	for {
		select {
		case <-self.quiting:
			return
		default:
			// pair should exist when there're responses
			io.Copy(self.pair.local, self.pair.remote)
			/*
				case data := <-self.outgoing:
					if _, err := self.writer.Write(data); err != nil {
						log.Critical("Write error: %s", err)
						self.quit()
					}

					if err := self.writer.Flush(); err != nil {
						log.Critical("Flush error: %s", err)
						self.quit()
					}
			*/
		}

	}

}

func (self *Connection) Close() {
	self.pair.local.Close()
	self.pair.remote.Close()
}

// Methods for Proxy

func CreateProxy() *Proxy {
	proxy := &Proxy{
		clients:  make(ConnectionTable),
		pending:  make(chan net.Conn),
		quiting:  make(chan net.Conn),
		incoming: make(chan TempData),
		connPair: make(chan ConnPair),
		//outgoing: make(chan []byte),
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
			case pair := <-self.connPair:
				self.clients[pair.local].pair.remote = pair.remote
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
			case data := <-connection.incoming:
				log.Info("Got temp data from client %v\n", connection)
				tmpData := TempData{
					conn: conn,
					data: data,
				}
				self.incoming <- tmpData
			}
		}
	}()
}

func (self *Proxy) cleanup(conn net.Conn) {
	if conn != nil {
		connection := self.clients[conn]
		connection.Close()
		delete(self.clients, conn)
	}
}

func (self *Proxy) Start(connString string) {
	self.listener, _ = net.Listen("tcp", connString)

	log.Notice("Proxy server %p starts\n", self)

	for {
		conn, err := self.listener.Accept()

		if err != nil {
			log.Error("Cannot accept web connection:", err)
			continue
		}

		log.Info("A new connection %v kicks\n", conn)
		self.pending <- conn
	}
}
