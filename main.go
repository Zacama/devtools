package main

import (
	"fmt"
	"strings"

	"devtools/backend"
	"devtools/net/ssh"
)

func main() {
	c, err := ssh.Connect("10.4.69.31", "22", "root", "152486379")
	if err != nil {
		panic(err)
	}

	clientSecretGet, err := c.Exec("kubectl exec svc/webconsole -n anyshare -c webconsole -- env | grep CLIENT_SECRET")
	if err != nil {
		panic(err)
	}
	clientSecret := strings.TrimSpace(strings.SplitN(string(clientSecretGet), "=", 2)[1])

	clientIDGet, err := c.Exec("kubectl exec svc/webconsole -n anyshare -c webconsole -- env | grep CLIENT_ID")
	if err != nil {
		panic(err)
	}
	clientID := strings.TrimSpace(strings.SplitN(string(clientIDGet), "=", 2)[1])

	auth := backend.NewOAuth("10.4.69.31", clientID, clientSecret, "266c6a42-6131-4d62-8f39-853e7093701c")
	token, err := auth.GetToken()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(token)
}
