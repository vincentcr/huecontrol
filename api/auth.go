package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/vincentcr/huecontrol/services"
)

func mustAuthenticate(h handler) handler {
	return func(c *HCContext, w http.ResponseWriter, r *http.Request) error {
		_, ok := c.GetUser()
		if !ok {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Please enter your username and password\"")
			return NewHttpError(http.StatusUnauthorized)
		}
		h(c, w, r)
		return nil
	}
}

type AuthMethod string
type AuthCreds []string

var (
	AuthMethodNone  AuthMethod = ""
	AuthMethodBasic AuthMethod = "Basic"
	AuthMethodToken AuthMethod = "Bearer"
)

func authenticate(c *HCContext, w http.ResponseWriter, r *http.Request) error {
	verify := func(method AuthMethod, creds AuthCreds) (services.User, error) {
		switch method {
		case AuthMethodBasic:
			username := creds[0]
			password := creds[1]
			return c.Services.Users.AuthenticateWithPassword(username, password)
		case AuthMethodToken:
			token := creds[0]
			return c.Services.Users.AuthenticateWithToken(token)
		default:
			return services.User{}, fmt.Errorf("Unknown auth method %v", method)
		}
	}

	user, err := authenticateRequest(verify, w, r)

	if err == noAuthAttempted {
		return nil
	} else if err == nil {
		c.Env["user"] = user
		log.Printf("Authenticated as %v", user)
	}
	return err
}

var noAuthAttempted = fmt.Errorf("no_auth_attempted")

type authVerification func(method AuthMethod, creds AuthCreds) (services.User, error)

func authenticateRequest(verify authVerification, w http.ResponseWriter, r *http.Request) (services.User, error) {

	method, creds, err := parseAuthorizationFromRequest(r)

	if err != nil {
		log.Println("Failed to parse credentials:", err)
		return services.User{}, NewHttpError(http.StatusBadRequest)
	}
	if method == AuthMethodNone {
		return services.User{}, err
	}

	user, err := verify(method, creds)
	if err == services.ErrNotFound {
		return services.User{}, NewHttpErrorWithText(http.StatusUnauthorized, "Invalid Credentials")
	} else if err != nil {
		return services.User{}, err
	} else {
		return user, nil
	}

}

type credentialParser func(r *http.Request) (AuthMethod, AuthCreds, error)

var credentialParsers = []credentialParser{parseAuthorizationFromHeader, parseAuthorizationFromForm}

func parseAuthorizationFromRequest(r *http.Request) (AuthMethod, AuthCreds, error) {

	for _, parser := range credentialParsers {
		method, creds, err := parser(r)
		if method != AuthMethodNone || err != nil {
			return method, creds, err
		}
	}
	return AuthMethodNone, nil, nil
}

func parseAuthorizationFromHeader(r *http.Request) (AuthMethod, AuthCreds, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return AuthMethodNone, nil, nil
	}

	match := regexp.MustCompile("^(.+?)\\s+(.+)$").FindStringSubmatch(header)
	if len(match) == 0 {
		return "", nil, fmt.Errorf("Invalid auth header")
	}

	method := AuthMethod(match[1])
	encodedCreds := match[2]
	var creds AuthCreds

	if method == AuthMethodBasic {
		userPasswordStr, err := base64.StdEncoding.DecodeString(encodedCreds)
		if err != nil {
			return "", nil, fmt.Errorf("Invalid basic auth header: not base64")
		}

		creds = strings.Split(string(userPasswordStr), ":")

	} else if method == AuthMethodToken {
		creds = []string{encodedCreds}
	}

	return method, creds, nil
}

func parseAuthorizationFromForm(r *http.Request) (AuthMethod, AuthCreds, error) {
	token := r.FormValue("_auth_token")
	if token != "" {
		return AuthMethodToken, []string{token}, nil
	}
	return AuthMethodNone, nil, nil
}
