package main

import (
	"devtools/backend"
	"fmt"
	"strings"

	"devtools/net/ssh"
)

func main() {
	host := "10.4.69.31"
	sshPort := "22"
	user := "root"
	password := "152486379"

	c, err := ssh.Connect(host, sshPort, user, password)
	if err != nil {
		panic(err)
	}

	/*
		kubectl get deployments.apps -n anyshare webconsole -o go-template='{{range.spec.template.spec.containers}}{{if eq .name "webconsole"}}{{range .env}}{{if eq .name "CLIENT_SECRET"}}{{.value}}{{end}}{{end}}{{end}}{{end}}'
		kubectl get deployments.apps -n anyshare webconsole -o go-template='{{range.spec.template.spec.containers}}{{if eq .name "webconsole"}}{{range .env}}{{if eq .name "CLIENT_ID"}}{{.value}}{{end}}{{end}}{{end}}{{end}}'
	*/
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

	userID, err := c.Exec("set -e; kubectl exec -n resource mariadb-mariadb-0 -c mariadb -- mysql -uroot -peisoo.com123 -e \"SELECT f_user_id FROM sharemgnt_db.t_user WHERE f_login_name = 'admin';\" | awk 'NR==2'")
	if err != nil {
		panic(err)
	}
	adminID := strings.TrimSpace(string(userID))

	auth := backend.NewOAuth(host, clientID, clientSecret, adminID)
	token, err := auth.GetToken()
	if err != nil {
		panic(err)
	}

	fmt.Println(token)
}
