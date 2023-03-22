package backend

import (
	"errors"
	"fmt"
	"github.com/imroc/req/v3"
	"net"
	"net/http"
	"strconv"
)

var (
	RedirectURL = func(ip net.IP) string { return fmt.Sprintf("https://%s:443/oauthlogin", ip.To4().String()) }
)

type ConsoleTokenGet interface {
	GetToken() (token string, err error)
}

type OAuth struct {
	Host         net.IP
	ClientID     string
	ClientSecret string
	UserID       string
	Cookie       *http.Cookie
}

func NewOAuth(host, clientID, clientSecret, userID string) OAuth {
	return OAuth{
		Host:         net.ParseIP(host).To4(),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		UserID:       userID,
	}
}

func (a *OAuth) GetToken() (token string, err error) {
	code, err := a.getCode()
	if err != nil {
		return "", err
	}

	data := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": RedirectURL(a.Host),
	}

	resp, err := req.C().
		SetRedirectPolicy(req.NoRedirectPolicy()).
		EnableInsecureSkipVerify().
		SetCommonBasicAuth(a.ClientID, a.ClientSecret).
		R().
		SetContentType("application/x-www-form-urlencoded").
		SetBody(data).
		Post(fmt.Sprintf("https://%s:443/oauth2/token", a.Host.String()))

	if err != nil {
		return "", err
	}

	if resp.GetStatusCode() != http.StatusOK {
		return "", errors.New(fmt.Sprintf("申请token失败, 状态码: %d, 响应体: %s", resp.GetStatusCode(), resp.String()))
	}

	return Response{resp}.GetJsonStringField("access_token"), nil
}

func (a *OAuth) getCode() (string, error) {
	redirectTo, err := a.acceptAuthRequest()
	if err != nil {
		return "", err
	}
	unquoteRedirectTo, err := strconv.Unquote(redirectTo)
	if err != nil {
		return "", err
	}

	resp, err := req.C().
		SetRedirectPolicy(req.NoRedirectPolicy()).
		EnableInsecureSkipVerify().
		R().
		SetCookies(a.Cookie).
		Get(unquoteRedirectTo)

	if resp.GetStatusCode() != http.StatusFound {
		return "", errors.New(fmt.Sprintf("getCode failed 状态码: %d, 响应体: %s", resp.GetStatusCode(), resp.String()))
	}

	// TODO
	resp.GetHeader("Location")

	return "", nil
}

func (a *OAuth) acceptAuthRequest() (string, error) {

	return "", nil
}
