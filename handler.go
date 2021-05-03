package main

import (
	"context"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

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
	param := r.URL.Query()
	if accessOrigin != "" {
		w.Header().Add("Access-Control-Allow-Origin", accessOrigin)
	}
	if _, ok := param["subscribe"]; ok {
		subscribe(w, r)
	} else if _, ok = param["unsubscribe"]; ok {
		unsubscribe(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("resources not found"))
	}
}

func subscribe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("param token is nil"))
		return
	}
	client, err := newClient()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("can't create messaging client"))
		return
	}
	_, err = client.SubscribeToTopic(context.Background(), []string{token}, "all")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("can't subscribe topic"))
		return
	}
	w.Write([]byte("subscribe successfully"))
}

func unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("param token is nil"))
		return
	}
	client, err := newClient()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("can't create messaging client"))
		return
	}
	_, err = client.UnsubscribeFromTopic(context.Background(), []string{token}, "all")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("can't unsubscribe topic"))
		return
	}
	w.Write([]byte("unsubscribe successfully"))
}
