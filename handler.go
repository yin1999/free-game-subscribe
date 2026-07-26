package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type payload struct {
	Method string `json:"method"`
	Token  string `json:"token"`
}

type response struct {
	Message string `json:"message"`
}

var (
	messagingClient   atomic.Pointer[messaging.Client]
	messagingClientMu sync.Mutex
)

func newClient() (client *messaging.Client, err error) {
	client = messagingClient.Load()
	if client != nil {
		return
	}
	messagingClientMu.Lock()
	defer messagingClientMu.Unlock()

	if client = messagingClient.Load(); client != nil {
		return
	}

	ctx := context.Background()
	opt := option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(os.Getenv("firebaseadminsdk")))
	var app *firebase.App
	app, err = firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return
	}
	client, err = app.Messaging(ctx)
	if err != nil {
		return
	}
	messagingClient.Store(client)
	return
}

func setCORSHeaders(w http.ResponseWriter, origin string) bool {
	if origin == "" {
		return false
	}

	_, originAllowed := accessOrigins[origin]
	_, wildcardAllowed := accessOrigins["*"]
	if !originAllowed && !wildcardAllowed {
		return false
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Add("Vary", "Origin")
	return true
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	corsAllowed := setCORSHeaders(w, r.Header.Get("Origin"))
	if r.Method == http.MethodOptions { // CORS
		if !corsAllowed {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		header := w.Header()
		header.Set("Access-Control-Allow-Headers", "Content-Type")
		header.Set("Access-Control-Allow-Methods", http.MethodPost)
		header.Set("Access-Control-Max-Age", "86400")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method == http.MethodPost {
		p := &payload{}
		if err := json.NewDecoder(r.Body).Decode(p); err != nil || p.Token == "" {
			writeJSON(w, &response{
				Message: "bad json format or empty token",
			}, http.StatusBadRequest)
			return
		}
		switch p.Method {
		case "subscribe", "unsubscribe":
			subscribeManage(r.Context(), w, p)
		default:
			writeJSON(w, &response{
				Message: "unsupport method",
			}, http.StatusNotAcceptable)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, data any, status ...int) {
	w.Header().Set("Content-Type", "application/json")
	if len(status) != 0 {
		w.WriteHeader(status[0])
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}

func subscribeManage(ctx context.Context, w http.ResponseWriter, p *payload) {
	client, err := newClient()
	if err != nil {
		writeJSON(w, &response{
			Message: "can't create messaging client",
		}, http.StatusInternalServerError)
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	switch p.Method {
	case "subscribe":
		_, err = client.SubscribeToTopic(ctx, []string{p.Token}, "all")
	case "unsubscribe":
		_, err = client.UnsubscribeFromTopic(ctx, []string{p.Token}, "all")
	}

	if err != nil {
		writeJSON(w, &response{
			Message: p.Method + " failed",
		}, http.StatusInternalServerError)
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	writeJSON(w, &response{
		Message: p.Method + " successfully",
	})
}
