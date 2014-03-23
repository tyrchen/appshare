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

type struct Control 