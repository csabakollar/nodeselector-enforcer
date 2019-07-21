package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/csabakollar/nodeselector-enforcer"
)

//main starts a local webserver to test the cloud function
func main() {
	http.HandleFunc("/", nodeselector.EntryPoint)
	fmt.Println("Listening on port 80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatalf("Cannot start HTTP server: %v", err)
	}
}
