package main

import (
	"flag"
	"log"

	"github.com/apt4105/journal/server"
)

var root = flag.String("root", ".", "root of the file schema")
var port = flag.String("port", ":8080", "the port to serve on")

func main() {
	srv := server.NewServer(*root, *port)

	err := srv.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
