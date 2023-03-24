package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const DefaultPort = "9999"

func getServerPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return port
	}

	return DefaultPort
}

// EchoHandler echos back the request as a response
func EchoHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Echoing back request made to " + request.URL.Path + " to client (" + request.RemoteAddr + ")")

	writer.Header().Set("Access-Control-Allow-Origin", "*")

	// allow pre-flight headers
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Range, Content-Disposition, Content-Type, ETag")

    // name responding hostname
    name, err := os.Hostname()
    if err != nil {
        panic(err)
    }
    writer.Header().Set("X-Responder", name)

    writer.Write([]byte(fmt.Sprintf("response from %s\n", name)))

	request.Write(writer)
}

func main() {
	log.Println("starting server, listening on port " + getServerPort())

	http.HandleFunc("/", EchoHandler)
	http.ListenAndServe(":"+getServerPort(), nil)
}
