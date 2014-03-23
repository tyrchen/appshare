package appshare

import (
	"net"
)

type Token chan int

// CloudProxy is the data structure to provide the entire context for cloud
// proxy.
type CloudProxy struct {
	control net.Listener // client listener for control messages
	data    net.Listener // client listener for data messages

	tokens Token // maximum local proxy can be connected to

}
