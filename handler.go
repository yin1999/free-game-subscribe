package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

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

func newClient() (client *messaging.Client, err error) {
	ctx := context.Background()
	opt := option.WithCredentialsJSON([]byte(os.Getenv("firebaseadminsdk")))
	var app *firebase.App
	app, err = firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return
	}
	client, err = app.Messaging(ctx)
	return
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if accessOrigin != "" {
		w.Header().Add("Access-Control-Allow-Origin", accessOrigin)
	}
	if r.Method == http.MethodOptions { // CORS
		header := w.Header()
		header.Add("Access-Control-Allow-Headers", "Content-Type")
		header.Add("Access-Control-Allow-Methods", http.MethodPost)
		header.Add("Access-Control-Max-Age", "86400")
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
			subscribeManage(w, p)
		default:
			writeJSON(w, &response{
				Message: "unsupport method",
			}, http.StatusNotAcceptable)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, data interface{}, status ...int) {
	w.Header().Add("Content-Type", "application/json")
	if len(status) != 0 {
		w.WriteHeader(status[0])
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}

func subscribeManage(w http.ResponseWriter, p *payload) {
	client, err := newClient()
	if err != nil {
		writeJSON(w, &response{
			Message: "can't create messaging client",
		}, http.StatusInternalServerError)
		return
	}

	switch p.Method {
	case "subscribe":
		_, err = client.SubscribeToTopic(context.Background(), []string{p.Token}, "all")
	case "unsubscribe":
		_, err = client.UnsubscribeFromTopic(context.Background(), []string{p.Token}, "all")
	}

	if err != nil {
		writeJSON(w, &response{
			Message: p.Method + " failed",
		}, http.StatusInternalServerError)
		return
	}
	writeJSON(w, &response{
		Message: p.Method + " successfully",
	})
}
