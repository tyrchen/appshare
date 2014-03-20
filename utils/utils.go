package utils

import (
	"io"
	"net"
	"regexp"
)

func Forward(local net.Conn, remote net.Conn) {
	go io.Copy(local, remote)
	go io.Copy(remote, local)
}

func Replace(regex string, data, partial []byte) []byte {
	reg, err := regexp.Compile(regex)
	if err != nil {
		panic(err)
	}
	return reg.ReplaceAll(data, partial)
}
