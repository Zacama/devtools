package backend

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

type pvtResponse struct {
	*req.Response
}

func (resp pvtResponse) GetJsonStringField(path string) string {
	result := gjson.Get(resp.String(), path)
	return result.String()
}

var (
	redirectURL = func(ip net.IP) string { return fmt.Sprintf("https://%s:443/oauthlogin", ip.To4().String()) }
	generateErr = func(method string, resp *req.Response) error {
		return errors.New(fmt.Sprintf("%s failed 状态码: %d, 响应体: %s", method, resp.GetStatusCode(), resp.String()))
	}
)

type ConsoleTokenGet interface {
	GetToken() (token string, err error)
}

type OAuth struct {
	Host         net.IP
	ClientID     string
	ClientSecret string
	UserID       string
	Cookies      []*http.Cookie
}

func NewOAuth(host, clientID, clientSecret, userID string) *OAuth {
	return &OAuth{
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
		"redirect_uri": redirectURL(a.Host),
	}

	resp, err := req.C().
		SetRedirectPolicy(req.NoRedirectPolicy()).
		EnableInsecureSkipVerify().
		SetCommonBasicAuth(a.ClientID, a.ClientSecret).
		R().
		SetFormData(data).
		Post(fmt.Sprintf("https://%s:443/oauth2/token", a.hostToString()))

	if err != nil {
		return "", err
	}

	if resp.GetStatusCode() != http.StatusOK {
		return "", generateErr("GetToken", resp)
	}

	return pvtResponse{resp}.GetJsonStringField("access_token"), nil
}

func (a *OAuth) getCode() (string, error) {
	redirectTo, err := a.acceptAuthRequest()
	if err != nil {
		return "", err
	}

	resp, err := req.C().
		SetRedirectPolicy(req.NoRedirectPolicy()).
		EnableInsecureSkipVerify().
		R().
		SetCookies(a.Cookies...).
		Get(redirectTo)
	if err != nil {
		return "", err
	}

	if resp.GetStatusCode() != http.StatusFound {

		return "", generateErr("getCode", resp)
	}

	locationStr := resp.GetHeader("Location")
	firstPart := strings.SplitN(locationStr, "&", 2)[0]
	firstPartSlice := strings.Split(firstPart, "=")
	code := firstPartSlice[len(firstPartSlice)-1]
	return code, nil
}

func (a *OAuth) acceptAuthRequest() (string, error) {
	// TODO 不必要的返回值
	_, consentChallenge, err := a.getAuthRequest()
	if err != nil {
		return "", err
	}

	params := map[string]string{"consent_challenge": consentChallenge}
	body := map[string]interface{}{
		"grant_access_token_audience": []interface{}{},
		"grant_scope":                 []string{"offline", "openid", "all"},
		"remember":                    false,
		"remember_for":                0,
		"session": map[string]interface{}{
			"access_token": map[string]interface{}{
				"account_type": "other",
				"client_type":  "console_web",
				"login_ip":     a.hostToString(),
				"udid":         "",
				"visitor_type": "realname",
			},
		},
	}

	resp, err := req.C().
		SetRedirectPolicy(req.NoRedirectPolicy()).
		EnableInsecureSkipVerify().
		R().
		SetQueryParams(params).
		SetCookies(a.Cookies...).
		SetBody(body).
		Put(fmt.Sprintf("https://%s:9080/oauth2/auth/requests/consent/accept", a.hostToString()))
	if err != nil {
		return "", err
	}

	if resp.GetStatusCode() != 200 {
		return "", generateErr("acceptLoginRequest", resp)
	}

	return pvtResponse{resp}.GetJsonStringField("redirect_to"), nil
}

func (a *OAuth) getAuthRequest() (interface{}, string, error) {
	consentChallenge, err := a.getAuthRedirect()
	if err != nil {
		return nil, "", err
	}
	params := map[string]string{"consent_challenge": consentChallenge}
	resp, err := req.C().
		SetRedirectPolicy(req.NoRedirectPolicy()).
		EnableInsecureSkipVerify().
		R().
		SetQueryParams(params).
		SetCookies(a.Cookies...).
		Get(fmt.Sprintf("https://%s:9080/oauth2/auth/requests/consent", a.hostToString()))
	if err != nil {
		return nil, "", err
	}
	if resp.GetStatusCode() != http.StatusOK {
		return nil, "", generateErr("getAuthRequest", resp)
	}

	body := make(map[string]interface{})
	err = resp.Unmarshal(&body)
	if err != nil {
		return nil, "", err
	}

	return body, consentChallenge, nil
}

func (a *OAuth) getAuthRedirect() (string, error) {
	redirectTo, err := a.acceptLoginRequest()
	if err != nil {
		return "", err
	}

	resp, err := req.C().
		SetRedirectPolicy(req.NoRedirectPolicy()).
		EnableInsecureSkipVerify().
		R().
		SetCookies(a.Cookies...).
		Get(redirectTo)
	if err != nil {
		return "", err
	}
	if resp.GetStatusCode() != http.StatusFound {
		return "", generateErr("getAuthRedirect", resp)
	}

	a.Cookies = resp.Cookies()
	s := strings.Split(resp.GetHeader("Location"), "=")
	return s[len(s)-1], nil
}

func (a *OAuth) acceptLoginRequest() (string, error) {
	challenge, err := a.getLoginRequest()
	if err != nil {
		return "", err
	}
	params := map[string]string{"login_challenge": challenge}
	body := map[string]interface{}{
		"acr": "string",
		"context": map[string]string{
			"account_type": "other",
			"client_type":  "console_web",
			"login_ip":     a.hostToString(),
			"udid":         "",
			"visitor_type": "realname",
		},
		"remember":     true,
		"remember_for": 3600,
		"subject":      a.UserID,
	}
	resp, err := req.C().
		SetRedirectPolicy(req.NoRedirectPolicy()).
		EnableInsecureSkipVerify().
		R().
		SetQueryParams(params).
		SetCookies(a.Cookies...).
		SetBody(body).
		Put(fmt.Sprintf("https://%s:9080/oauth2/auth/requests/login/accept", a.hostToString()))
	if err != nil {
		return "", err
	}
	if resp.GetStatusCode() != 200 {
		return "", generateErr("acceptLoginRequest", resp)
	}

	return pvtResponse{resp}.GetJsonStringField("redirect_to"), nil
}

func (a *OAuth) getLoginRequest() (string, error) {
	code, err := a.authorizingUsers()
	if err != nil {
		return "", err
	}
	params := map[string]string{"login_challenge": code}

	resp, err := req.C().
		SetRedirectPolicy(req.NoRedirectPolicy()).
		EnableInsecureSkipVerify().
		R().
		SetQueryParams(params).
		SetCookies(a.Cookies...).
		Get(fmt.Sprintf("https://%s:9080/oauth2/auth/requests/login", a.hostToString()))
	if err != nil {
		return "", err
	}
	if resp.GetStatusCode() != 200 {
		return "", generateErr("getLoginRequest", resp)
	}

	return pvtResponse{resp}.GetJsonStringField("challenge"), nil
}

func (a *OAuth) authorizingUsers() (string, error) {
	id, err := uuid.NewV1()
	if err != nil {
		return "", err
	}

	params := map[string]string{
		"audience":      "",
		"client_id":     a.ClientID,
		"response_type": "code",
		"scope":         "offline openid all",
		"redirect_uri":  redirectURL(a.Host),
		"state":         strings.Replace(id.String(), "-", "", -1),
	}

	resp, err := req.C().
		SetRedirectPolicy(req.NoRedirectPolicy()).
		EnableInsecureSkipVerify().
		R().
		SetQueryParams(params).
		Get(fmt.Sprintf("https://%s/oauth2/auth", a.hostToString()))
	if err != nil {
		return "", err
	}
	if resp.GetStatusCode() != http.StatusFound {
		return "", generateErr("authorizingUsers", resp)
	}

	a.Cookies = resp.Cookies()
	firstPart := strings.Split(resp.GetHeader("Location"), "=")
	return firstPart[len(firstPart)-1], nil
}

func (a *OAuth) hostToString() string {
	return a.Host.To4().String()
}
