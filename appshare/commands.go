package appshare

import (
	"errors"
	"github.com/tyrchen/goutil/regex"
	"regexp"
)

type Command struct {
	cmd string
	arg string
}

type Run func(server *Server, client *Client, arg string)

const (
	CMD_REGEX = `:(?P<cmd>\w+)\s*(?P<arg>.*)`
)

var (
	commands map[string]Run
)

func init() {
	commands = map[string]Run{
		"Name": setName,
		"Conn": createConnection,
	}
}

func parseCommand(msg string) (cmd Command, err error) {
	r := regexp.MustCompile(CMD_REGEX)
	if values, ok := regex.MatchAll(r, msg); ok {
		cmd.cmd, _ = values[0]["cmd"]
		cmd.arg, _ = values[0]["arg"]
		return
	}
	err = errors.New("Unparsed message: " + msg)
	return
}

func (self *Server) executeCommand(client *Client, cmd Command) (err error) {
	if f, ok := commands[cmd.cmd]; ok {
		f(self, client, cmd.arg)
		return
	}

	err = errors.New("Unsupported command: " + cmd.cmd)
	return
}

// commands

func setName(server *Server, client *Client, arg string) {
	//oldname := client.GetName()
	client.SetName(arg)
}

func createConnection(server *Server, client *Client, arg string) {
}
