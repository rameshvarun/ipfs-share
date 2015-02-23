package main

import (
	"os"
	"flag"
	"log"

	"github.com/go-martini/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/boltdb/bolt"
)

func main() {
	// Create content storage directories
	if err := os.MkdirAll("files/text", 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll("files/video", 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll("files/images", 0755); err != nil {
		log.Fatal(err)
	}

	//gatewayURL := flag.String("gateway", "http://gateway.ipfs.io", "The HTTP gateway used in shared file links.")
	flag.Parse()

	db, err := bolt.Open("blog.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Run()
}