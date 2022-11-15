package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/buildkite/agent/v3/api"
	"github.com/buildkite/agent/v3/logger"
)

func TestRegisteringAndConnectingClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/register":
			if got, want := authToken(req), "llamas"; got != want {
				http.Error(rw, fmt.Sprintf("authToken(req) = %q, want %q", got, want), http.StatusUnauthorized)
				return
			}
			rw.WriteHeader(http.StatusOK)
			fmt.Fprint(rw, `{"id":"12-34-56-78-91", "name":"agent-1", "access_token":"alpacas"}`)

		case "/connect":
			if got, want := authToken(req), "alpacas"; got != want {
				http.Error(rw, fmt.Sprintf("authToken(req) = %q, want %q", got, want), http.StatusUnauthorized)
				return
			}
			rw.WriteHeader(http.StatusOK)
			fmt.Fprint(rw, `{}`)

		default:
			http.Error(rw, fmt.Sprintf("not found; method = %q, path = %q", req.Method, req.URL.Path), http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Initial client with a registration token
	c := api.NewClient(logger.Discard, api.Config{
		Endpoint: server.URL,
		Token:    "llamas",
	})

	// Check a register works
	regResp, _, err := c.Register(&api.AgentRegisterRequest{})
	if err != nil {
		t.Fatalf("c.Register(&AgentRegisterRequest{}) error = %v", err)
	}

	if got, want := regResp.Name, "agent-1"; got != want {
		t.Errorf("regResp.Name = %q, want %q", got, want)
	}

	if got, want := regResp.AccessToken, "alpacas"; got != want {
		t.Errorf("regResp.AccessToken = %q, want %q", got, want)
	}

	// New client with the access token
	c2 := c.FromAgentRegisterResponse(regResp)

	// Check a connect works
	if _, err := c2.Connect(); err != nil {
		t.Errorf("c.FromAgentRegisterResponse(regResp).Connect() error = %v", err)
	}
}

func authToken(req *http.Request) string {
	return strings.TrimPrefix(req.Header.Get("Authorization"), "Token ")
}
