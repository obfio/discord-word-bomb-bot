package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func sendCommand(w http.ResponseWriter, r *http.Request) {
	c := &command{}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(b, &c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	httpCommandChannel <- c
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Command sent"))
}
