package appshare

import (
	"fmt"
	"github.com/op/go-logging"
	"net"
	"strings"
)

var (
	log = logging.MustGetLogger("appshare")
)

// ClouldContext contains all the information for cloud proxy to work.
type CloudContext struct {
	proxy    *Proxy       // proxy
	control  *Server      // control connections
	listener net.Listener // data connections
	// not good but I don't have better solutions
	local  chan net.Conn // temporarily store local connection
	remote chan net.Conn // temporarily store remote connection
}

// CloudContext methods

// CreateCloudContext creates a context for cloud proxy.
func CreateCloudContext() *CloudContext {
	proxy := CreateProxy()
	control := CreateServer()
	ctx := &CloudContext{
		proxy:   proxy,
		control: control,
	}
	return ctx
}

func (self *CloudContext) listen() {
	go func() {
		select {
		case tmpData := <-self.proxy.incoming: // request for create connection
			self.local <- tmpData.conn
			data := tmpData.data
			client := self.control.getClient(data)
			self.control.sendCommand(client, fmt.Sprintf(":conn %s"))
		case message := <-self.control.incoming:
			if strings.Contains(message, ":name") {
				local := <-self.local
				remote := <-self.remote
				self.proxy.clients[local].pair.remote = remote
			}
		}
	}()
}

// Start activates all listeners. User should pass a connection string to cloud
// proxy server, local proxy control, local proxy data.
func (self *CloudContext) Start(cpConnStr, lpCtrlConnStr, lpDataConnStr string) {
	self.listener, _ = net.Listen("tcp", lpDataConnStr)

	self.proxy.Start(cpConnStr)
	self.control.Start(lpCtrlConnStr)

	for {
		conn, err := self.listener.Accept()

		if err != nil {
			log.Error("Error on accepting data connection:", err)
			continue
		}

		log.Info("A new data connection %v kicks\n", conn)

		self.remote <- conn
	}
}
