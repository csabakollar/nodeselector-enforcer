package main

import (
	"fmt"
	"net/http"

	"github.com/csabakollar/nodeselector-enforcer"
)

//main starts a local webserver to test the cloud function
func main() {
	http.HandleFunc("/", nodeselector.EntryPoint)
	fmt.Println("Listening on port 80")
	http.ListenAndServe(":80", nil)
}
