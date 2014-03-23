package main

import (
	. "./appshare"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s <control port for client> <data port for client> <port for web server>\n", os.Args[0])
		os.Exit(-1)
	}

	ctx := CreateCloudContext()
	ctx.Start(os.Args[2], os.Args[3], os.Args[1])
}
