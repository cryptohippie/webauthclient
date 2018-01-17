package webauthclient

import (
	"bytes"
	"errors"
	html "html/template"
	"io"
	"net/http"
	"net/url"
	template "text/template"
)

var (
	// ErrAPI is returned if the API does not behave as expected.
	ErrAPI = errors.New("webauthclient: API incorrect")
	// ErrAuth is returned if we suspect an authentication error.
	ErrAuth = errors.New("webauthclient: Authentication error")
	// ErrSigned is returned if the authenticator does not accept the signed token.
	ErrSigned = errors.New("webauthclient: Authenticator refused token")
	// ErrUser is returned if clientID or password have not been given.
	ErrUser = errors.New("webauthclient: ClientID or Password emtpy")
)

var (
	// TokenQueryKey is the key to look up the token in the query.
	TokenQueryKey = "token"
)

const (
	locationHeaderString = "Location"
	signedURLIndicator   = "SIGNEDTOKEN:"
	defaultAuthURL       = "https://auth.cryptohippie.net/"
	postHeadersTemplate  = "form_name=login&token={{.Token}}&clientid={{.ClientID}}&password={{.Password}}"
	successString        = "Connection authenticated."
)

// Authenticator is a client authenticator for the CHAVPN authentication framework using the standard web login.
type Authenticator struct {
	// AuthURL is the URL at which the authentication server listens. If it is an empty string
	// it will be set to the default URL.
	AuthURL string
	// PostHeadersTemplate is the template that defines the post fields to the authenticator/login
	// server. If it is an empty string it will be set to the default PostHeadersTemplate.
	PostHeadersTemplate string
	// HTTPClientFactory is a function that will be called by the authenticator to create HTTP clients.
	// If it is set to nil, a default factory is used.
	HTTPClientFactory HTTPClientFactoryFunc
}

func defaultHTTPClientFactory() *http.Client {
	return &http.Client{}
}

// Authenticate ClientID/Password to the CHAVPN authentication framework.
// It returns nil on success, error if there's an error.
func (auth *Authenticator) Authenticate(clientID, password string) error {
	if clientID == "" || password == "" {
		return ErrUser
	}
	if auth == nil {
		auth = new(Authenticator)
	}
	if auth.AuthURL == "" {
		auth.AuthURL = defaultAuthURL
	}
	if auth.PostHeadersTemplate == "" {
		auth.PostHeadersTemplate = postHeadersTemplate
	}
	if auth.HTTPClientFactory == nil {
		auth.HTTPClientFactory = defaultHTTPClientFactory
	}
	redirURL, token, err := auth.getToken()
	if err != nil {
		return err
	}
	signURL, err := auth.getSignedToken(redirURL, token, clientID, password)
	if err != nil {
		return err
	}
	return auth.callAuthenticator(signURL)
}

// callAuthenticator calls the given URL as authenticator.
func (auth Authenticator) callAuthenticator(signURL string) error {
	client := auth.HTTPClientFactory()
	resp, err := client.Get(signURL)
	if err != nil {
		return err
	}
	respBuffer := bytes.NewBuffer(nil)
	io.CopyN(respBuffer, resp.Body, 1024*1024) // errors dont matter, content does.
	respBytes := respBuffer.Bytes()
	p := bytes.Index(respBytes, []byte(successString))
	if p < 0 {
		return ErrSigned
	}
	return nil
}

// getToken queries the authentication token from the authentication endpoint
// defined as authurl.
// The function returns the redirectURL, the extracted token, or an error.
func (auth Authenticator) getToken() (redirURL, token string, err error) {
	var tokens []string
	var ok bool
	if _, err := url.Parse(auth.AuthURL); err != nil || auth.AuthURL == "" {
		panic("Bad url in GetToken")
	}
	client := auth.HTTPClientFactory()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Get(auth.AuthURL)
	if err != nil {
		return "", "", err
	}
	resp.Body.Close()
	locationHeader := resp.Header.Get(locationHeaderString)
	if locationHeader == "" {
		return "", "", ErrAPI
	}
	locationURL, err := url.Parse(locationHeader)
	if err != nil {
		return "", "", ErrAPI
	}
	locationQuery := locationURL.Query()

	if tokens, ok = locationQuery[TokenQueryKey]; !ok {
		return "", "", ErrAPI
	}
	if len(tokens) < 1 {
		return "", "", ErrAPI
	}
	return locationURL.String(), tokens[0], nil
}

// URLFields contains fields used in post calls.
type URLFields struct {
	Token    string
	ClientID string
	Password string
}

// getSignedToken calls the redirect URL returned by GetToken to request a signed
// authentication URL. If the given http.Client is nil, a new one will be created.
func (auth Authenticator) getSignedToken(redirURL, token, clientID, password string) (signURL string, err error) {
	client := auth.HTTPClientFactory()
	postTmp := template.Must(template.New("post").Parse(auth.PostHeadersTemplate))
	urlFields := &URLFields{
		Token:    html.URLQueryEscaper(token),
		ClientID: html.URLQueryEscaper(clientID),
		Password: html.URLQueryEscaper(password),
	}
	buffer := bytes.NewBuffer(nil)
	err = postTmp.Execute(buffer, urlFields)
	if err != nil {
		return "", err
	}
	resp, err := client.Post(redirURL, "application/x-www-form-urlencoded", buffer)
	if err != nil {
		resp.Body.Close()
		return "", err
	}
	respBuffer := bytes.NewBuffer(nil)
	io.CopyN(respBuffer, resp.Body, 1024*1024) // errors dont matter, content does.
	respBytes := respBuffer.Bytes()
	p := bytes.Index(respBytes, []byte(signedURLIndicator))
	if p < 0 {
		return "", ErrAuth
	}
	e := bytes.IndexByte(respBytes[p:], ' ')
	if p < 0 {
		return "", ErrAPI
	}
	return string(respBytes[p+len(signedURLIndicator) : p+e]), nil
}
