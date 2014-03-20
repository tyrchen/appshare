package main

import (
	"./utils"
	"fmt"
	"log"
	"net"
	"os"
)

func listenToClients(connStr string) net.Conn {
	listener, _ := net.Listen("tcp", connStr)

	log.Printf("Listening to clients\n")

	conn, err := listener.Accept()

	if err != nil {
		panic(err)
	}

	log.Printf("A client connection %v kicks\n", conn)

	return conn
}

func listenToWebServer(connStr string, remote net.Conn) {
	listener, _ := net.Listen("tcp", connStr)

	log.Printf("Listening to webserver\n")

	for {
		local, err := listener.Accept()

		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("A web server connection %v kicks\n", local)

		utils.Forward(local, remote)
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <port for client> <port for web server>\n", os.Args[0])
		os.Exit(-1)
	}

	remote := listenToClients(os.Args[1])
	listenToWebServer(os.Args[2], remote)

}
