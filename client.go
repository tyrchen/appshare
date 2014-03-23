package main

import (
	"./utils"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	replacement string
)

func dialToServer(connStr string) net.Conn {
	conn, err := net.Dial("tcp", connStr)

	log.Printf("Dial to server: %v\n", conn)

	if err != nil {
		panic(err)
	}

	return conn
}

func httpProxy(local, remote net.Conn, data []byte) {
	go local.Write(data)
	go io.Copy(remote, local)

}

func replaceHost(data []byte) []byte {
	regex := `Host: [^\r]+`
	partial := []byte(replacement)

	return utils.Replace(regex, data, partial)
}

func dialToWebServer(connStr string, remote net.Conn) {
	bytes := make([]byte, 30000)
	for {
		n, err := remote.Read(bytes)
		if err != nil {
			panic(err)
		}
		bytes := replaceHost(bytes)
		data := string(bytes[:n])
		fmt.Println("Request bytes: ", n)
		fmt.Println(data)

		local, err := net.Dial("tcp", connStr)

		log.Printf("Dial to web server: %v\n", local)

		if err != nil {
			panic(err)
		}

		httpProxy(local, remote, bytes[:n])
	}
}

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s <control port for server> <data port for server> <port for web server>\n", os.Args[0])
		os.Exit(-1)
	}

	replacement = fmt.Sprintf("Host: localhost%s", os.Args[2])
	remote := dialToServer(os.Args[1])
	dialToWebServer(os.Args[2], remote)

}
