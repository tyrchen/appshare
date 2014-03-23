package utils

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/tyrchen/goutil/uniq"
	"io"
	"net"
	"regexp"
)

const (
	NAME_PREFIX = "appshare"
)

func GetUniqName() string {
	return fmt.Sprintf("%s-%d", NAME_PREFIX, uniq.GetUniq())
}

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

func Hash(content string) string {
	h := sha1.New()
	io.WriteString(h, content)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func HashEqual(content string, sha1base64 string) bool {
	return Hash(content) == sha1base64
}
