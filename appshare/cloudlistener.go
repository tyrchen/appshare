package appshare

import (
	"../utils"
	"bufio"
	"net"
)

const (
	MAXCLIENTS = 50
)

type Message chan string
type Token chan int

type Client struct {
	conn     net.Conn
	incoming Message
	outgoing Message
	reader   *bufio.Reader
	writer   *bufio.Writer
	quiting  chan net.Conn
	name     string
}

type ClientTable map[net.Conn]*Client
type ClientNameTable map[string]*Client

type Server struct {
	listener net.Listener
	clients  ClientTable
	names    ClientNameTable
	tokens   Token
	pending  chan net.Conn
	quiting  chan net.Conn
	incoming Message
	outgoing Message
}

// Client methods
func (self *Client) GetName() string {
	return self.name
}

func (self *Client) SetName(name string) {
	self.name = name
}

func (self *Client) GetIncoming() string {
	return <-self.incoming
}

func (self *Client) PutOutgoing(message string) {
	self.outgoing <- message
}

func CreateClient(conn net.Conn) *Client {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	client := &Client{
		conn:     conn,
		incoming: make(Message),
		outgoing: make(Message),
		quiting:  make(chan net.Conn),
		reader:   reader,
		writer:   writer,
	}
	client.Listen()
	return client
}

func (self *Client) Listen() {
	go self.Read()
	go self.Write()
}

func (self *Client) quit() {
	close(self.quiting)
}

func (self *Client) Read() {
	select {
	case <-self.quiting:
		return
	default:
		if line, _, err := self.reader.ReadLine(); err == nil {
			self.incoming <- string(line)
		} else {
			log.Error("Read error: %s", err)
			self.quit()
			return
		}
	}

}

func (self *Client) Write() {
	select {
	case <-self.quiting:
		return
	case data := <-self.outgoing:
		if _, err := self.writer.WriteString(data + "\n"); err != nil {
			log.Error("Write error: %s\n", err)
			self.quit()
			return
		}

		if err := self.writer.Flush(); err != nil {
			log.Error("Flush error: %s\n", err)
			self.quit()
			return
		}
	}
}

func (self *Client) Close() {
	self.conn.Close()
}

// Server methods
func (self *Server) generateToken() {
	self.tokens <- 0
}

func (self *Server) takeToken() {
	<-self.tokens
}

func CreateServer() *Server {
	server := &Server{
		clients:  make(ClientTable, MAXCLIENTS),
		names:    make(ClientNameTable, MAXCLIENTS),
		tokens:   make(Token, MAXCLIENTS),
		pending:  make(chan net.Conn),
		quiting:  make(chan net.Conn),
		incoming: make(Message),
		outgoing: make(Message),
	}
	server.listen()
	return server
}

func (self *Server) listen() {
	go func() {
		for {
			select {
			case conn := <-self.pending:
				self.join(conn)
			case conn := <-self.quiting:
				self.leave(conn)
			}
		}
	}()
}

func (self *Server) join(conn net.Conn) {
	client := CreateClient(conn)
	name := utils.GetUniqName()
	client.SetName(name)
	self.clients[conn] = client
	self.names[name] = client

	log.Info("Auto assigned name for conn %p: %s\n", conn, name)

	go func() {
		for {
			select {
			case conn := <-client.quiting:
				self.quiting <- conn
				return

			case msg := <-client.incoming:
				log.Info("Got message: %s from client %s\n", msg, client.GetName())
				self.incoming <- msg

				/* Should be run in the client proxy
				if strings.HasPrefix(msg, ":") {
					if cmd, err := parseCommand(msg); err == nil {
						if err = self.executeCommand(client, cmd); err == nil {
							continue
						} else {
							log.Warning(err.Error())
						}
					} else {
						log.Warning(err.Error())
					}
				}
				*/
			}
		}
	}()
}

func (self *Server) leave(conn net.Conn) {
	if conn != nil {
		conn.Close()
		client := self.clients[conn]
		delete(self.clients, conn)
		delete(self.names, client.name)
	}

	self.generateToken()
}

func (self *Server) Start(connString string) {
	self.listener, _ = net.Listen("tcp", connString)

	log.Notice("Server %p starts\n", self)

	// filling the tokens
	for i := 0; i < MAXCLIENTS; i++ {
		self.generateToken()
	}

	for {
		conn, err := self.listener.Accept()

		if err != nil {
			log.Error("Cannot accept control connection:", err)
			continue
		}

		log.Info("A new control connection %v kicks\n", conn)

		self.takeToken()
		self.pending <- conn
	}
}

func (self *Server) getClient(data []byte) *Client {
	// temporary code for testing purpose
	for _, v := range self.clients {
		return v
	}
	return nil
}

func (self *Server) sendCommand(client *Client, command string) {
	client.PutOutgoing(command)
}
