package mitch

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type response struct {
	s      *server
	w      http.ResponseWriter
	req    *http.Request
	status int

	currentUser *User
}

type Any map[string]interface{}

type APIError struct {
	status   int
	messages []string
}

func Error(status int, messages ...string) APIError {
	return APIError{
		status:   status,
		messages: messages,
	}
}

func Throw(status int, messages ...string) APIError {
	panic(Error(status, messages...))
}

func (ae APIError) Error() string {
	return fmt.Sprintf("api error (%d): %v", ae.status, ae.messages)
}

func (r *response) WriteError(status int, errors ...string) {
	r.status = status
	payload := map[string]interface{}{
		"errors": errors,
	}
	r.WriteJSON(payload)
}

func (r *response) WriteJSON(payload interface{}) {
	r.Header().Set("content-type", "application/json")
	r.WriteHeader()

	bs, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		panic(err)
	}

	r.Write(bs)
}

func (r *response) Header() http.Header {
	return r.w.Header()
}

func (r *response) WriteHeader() {
	status := r.status
	if r.status == 0 {
		status = 200
	}
	r.w.WriteHeader(status)
}

func (r *response) Write(p []byte) {
	r.w.Write(p)
}

func (r *response) CheckAPIKey() {
	keyString := r.req.Header.Get("Authorization")
	if keyString == "" {
		keyString = r.req.URL.Query().Get("api_key")
	}
	if keyString == "" {
		Throw(401, "authentication required")
	}

	apiKey := r.s.store.FindAPIKeysByKey(keyString)
	if apiKey == nil {
		Throw(403, "unauthorized")
	}

	r.currentUser = r.s.store.FindUser(apiKey.UserID)
	if r.currentUser == nil {
		Throw(500, "api key has no user")
	}
}
